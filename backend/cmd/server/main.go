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
	applicationHabit "github.com/Watari995/streek/backend/internal/application/habit"
	"github.com/Watari995/streek/backend/internal/config"
	domainService "github.com/Watari995/streek/backend/internal/domain/service"
	"github.com/Watari995/streek/backend/internal/handler"
	infraAuth "github.com/Watari995/streek/backend/internal/infrastructure/auth"
	"github.com/Watari995/streek/backend/internal/infrastructure/cache"
	"github.com/Watari995/streek/backend/internal/infrastructure/database"
	"github.com/Watari995/streek/backend/internal/middleware"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/bcrypt"
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

	// repository
	userRepo := database.NewUserRepository(db)
	habitRepo := database.NewHabitRepository(db)
	checkInRepo := database.NewCheckInRepository(db)
	hasher := infraAuth.NewBcryptHasher(bcrypt.DefaultCost)
	tokenGenerator := infraAuth.NewJWTGenerator([]byte(cfg.JWT.Secret))

	// domain services
	streakService := domainService.NewStreakService()

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
	checkInService := applicationCheckIn.NewCheckIn(checkInRepo, habitRepo)
	undoService := applicationCheckIn.NewUndo(checkInRepo, habitRepo)

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
	habits := api.Group("/habits", middleware.AuthMiddleware(tokenGenerator))
	habits.GET("", habitHandler.List)
	habits.POST("", habitHandler.Create)
	habits.PUT("/:id", habitHandler.Update)
	habits.DELETE("/:id", habitHandler.Delete)
	checkIns := habits.Group("/:id/check")
	checkIns.POST("", checkInHandler.CheckIn)
	checkIns.DELETE("", checkInHandler.Undo)

	// stats
	stats := api.Group("/stats", middleware.AuthMiddleware(tokenGenerator))
	// optional "today" query parameter
	stats.GET("/overview", statsHandler.GetOverview)

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
