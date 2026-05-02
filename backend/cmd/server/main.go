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
	"github.com/Watari995/streek/backend/internal/config"
	"github.com/Watari995/streek/backend/internal/handler"
	infraAuth "github.com/Watari995/streek/backend/internal/infrastructure/auth"
	"github.com/Watari995/streek/backend/internal/infrastructure/database"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	// user repository
	userRepo := database.NewUserRepository(db)
	// hasher
	hasher := infraAuth.NewBcryptHasher(bcrypt.DefaultCost)
	// token generator
	tokenGenerator := infraAuth.NewJWTGenerator([]byte(cfg.JWT.Secret))

	// services
	registerService := applicationAuth.NewRegister(userRepo, hasher, tokenGenerator)
	loginService := applicationAuth.NewLogin(userRepo, hasher, tokenGenerator)

	// auth handler
	authHandler := handler.NewAuthHandler(
		registerService,
		loginService,
	)

	// middleware settings
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	api := e.Group("/api/v1")
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)

	// start goroutine for background tasks
	go func() {
		if err := e.Start(cfg.Server.Port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			e.Logger.Fatal("shutting down the server")
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
