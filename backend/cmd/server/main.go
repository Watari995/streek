package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	// middleware settings
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// start goroutine for background tasks
	go func() {
		if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
