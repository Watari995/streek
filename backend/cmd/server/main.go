package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	applicationAuth "github.com/Watari995/streek/backend/internal/application/auth"
	applicationCheckIn "github.com/Watari995/streek/backend/internal/application/check_in"
	"github.com/Watari995/streek/backend/internal/application/eventhandler"
	applicationHabit "github.com/Watari995/streek/backend/internal/application/habit"
	applicationPoint "github.com/Watari995/streek/backend/internal/application/point"
	"github.com/Watari995/streek/backend/internal/config"
	"github.com/Watari995/streek/backend/internal/domain/event/types"
	domainService "github.com/Watari995/streek/backend/internal/domain/service"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
	"github.com/Watari995/streek/backend/internal/handler"
	infraAuth "github.com/Watari995/streek/backend/internal/infrastructure/auth"
	"github.com/Watari995/streek/backend/internal/infrastructure/cache"
	"github.com/Watari995/streek/backend/internal/infrastructure/circuitbreaker"
	"github.com/Watari995/streek/backend/internal/infrastructure/database"
	"github.com/Watari995/streek/backend/internal/infrastructure/event"
	"github.com/Watari995/streek/backend/internal/infrastructure/notification"
	"github.com/Watari995/streek/backend/internal/infrastructure/ratelimit"
	"github.com/Watari995/streek/backend/internal/middleware"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/samber/lo"
	"golang.org/x/crypto/bcrypt"
)

const (
	// naming?
	rateLimitLimit  = 10
	rateLimitWindow = 1 * time.Minute // 1 minute
)

func main() {
	e := echo.New()

	// load config
	cfg, err := config.Load()
	if err != nil {
		e.Logger.Fatal(err)
	}

	// database connection
	db, err := database.NewDB(cfg.DB.DSN())
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer db.Close()

	// Redis connection
	redisClient, err := cache.NewRedisClient(cfg.Redis)
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer redisClient.Close()
	streakCache := cache.NewStreakCache(redisClient)

	// rate limiter
	rateLimiter := ratelimit.NewRedisRateLimiter(redisClient, rateLimitLimit, rateLimitWindow, time.Now)

	// domain services
	streakService := domainService.NewStreakService()

	// repository
	userRepo := database.NewUserRepository(db)
	habitRepo := database.NewHabitRepository(db)
	checkInRepo := database.NewCheckInRepository(db)
	eventPublisher := event.NewInMemoryPublisher()
	pointLedgerRepo := database.NewPointLedgerRepository(db)
	txManager := database.NewTransactionManager(db)
	hasher := infraAuth.NewBcryptHasher(bcrypt.DefaultCost)
	tokenGenerator := infraAuth.NewJWTGenerator([]byte(cfg.JWT.Secret))
	// smtp notification with circuit breaker(repository layer)
	if cfg.Notification.IsSMTPEnabled() {
		emailNotifier := notification.NewEmailNotifier(cfg.Notification.SMTPHost, cfg.Notification.SMTPPort, cfg.Notification.SMTPUser, cfg.Notification.SMTPPassword, cfg.Notification.SMTPFrom)
		cb := circuitbreaker.New("smtp", 3, 30*time.Second)
		notifier := notification.NewCircuitBreakerNotifier(emailNotifier, cb)
		notifyTo := lo.Must(valueobject.NewEmail(cfg.Notification.To))
		notifyHandler := eventhandler.NewNotifyStreakMilestone(notifier, checkInRepo, streakService, notifyTo)
		eventPublisher.SubscribeAsync(types.EventTypeCheckInSucceeded, notifyHandler.Handle)
	}

	// handler
	earnPointsOnCheckInHandler := eventhandler.NewEarnPointsOnCheckIn(pointLedgerRepo)
	eventPublisher.Subscribe(types.EventTypeCheckInCompleted, earnPointsOnCheckInHandler.Handle)

	// services
	registerService := applicationAuth.NewRegister(userRepo, hasher, tokenGenerator)
	loginService := applicationAuth.NewLogin(userRepo, hasher, tokenGenerator)
	// habit
	listService := applicationHabit.NewList(habitRepo)
	createService := applicationHabit.NewCreate(habitRepo)
	updateService := applicationHabit.NewUpdate(habitRepo)
	deleteService := applicationHabit.NewDelete(habitRepo)
	getOverviewService := applicationHabit.NewGetOverview(habitRepo, checkInRepo, streakService, streakCache)
	// checkIn
	checkInService := applicationCheckIn.NewCheckIn(checkInRepo, habitRepo, streakCache, eventPublisher, txManager)
	undoService := applicationCheckIn.NewUndo(checkInRepo, habitRepo, streakCache)
	// point
	getBalanceService := applicationPoint.NewGetBalance(pointLedgerRepo)
	getHistoryService := applicationPoint.NewGetHistory(pointLedgerRepo)

	// auth handler
	authHandler := handler.NewAuthHandler(
		registerService,
		loginService,
	)
	habitHandler := handler.NewHabitHandler(
		listService,
		createService,
		updateService,
		deleteService,
	)
	checkInHandler := handler.NewCheckInHandler(
		checkInService,
		undoService,
	)
	statsHandler := handler.NewStatsHandler(getOverviewService)
	pointHandler := handler.NewPointHandler(getBalanceService, getHistoryService)

	// middleware settings
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())

	// health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// handler
	api := e.Group("/api/v1")
	// auth
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)

	// habits
	habits := api.Group("/habits", middleware.AuthMiddleware(tokenGenerator), middleware.RateLimitMiddleware(rateLimiter))
	habits.GET("", habitHandler.List)
	habits.POST("", habitHandler.Create)
	habits.PUT("/:id", habitHandler.Update)
	habits.DELETE("/:id", habitHandler.Delete)
	checkIns := habits.Group("/:id/check")
	checkIns.POST("", checkInHandler.CheckIn)
	checkIns.DELETE("", checkInHandler.Undo)

	// stats
	stats := api.Group("/stats", middleware.AuthMiddleware(tokenGenerator), middleware.RateLimitMiddleware(rateLimiter))
	// optional "today" query parameter
	stats.GET("/overview", statsHandler.GetOverview)

	// points
	points := api.Group("/points", middleware.AuthMiddleware(tokenGenerator), middleware.RateLimitMiddleware(rateLimiter))
	points.GET("/balance", pointHandler.GetBalance)
	points.GET("/history", pointHandler.GetHistory)

	// start goroutine for background tasks
	go func() {
		if err := e.Start(":" + cfg.Server.Port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			e.Logger.Fatal(err)
		}
	}()

	// SIGINTまたはSIGTERMを待つ
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	// in-flight リクエストの完了を最大10秒まつ
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
