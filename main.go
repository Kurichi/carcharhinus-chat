package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Kurichi/carcharhinus-chat/handler"
	"github.com/Kurichi/carcharhinus-chat/infra"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Setup
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())
	e.Use(middleware.Recover())

	ott := handler.NewOTTHandler()
	e.GET("/ott", ott.GenToken)

	repo := infra.NewRoomRepository()
	m := handler.NewChatHandler(repo)
	e.GET("/:roomID/ws", m.Join, ott.Middleware)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	// Start server
	go func() {
		if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
