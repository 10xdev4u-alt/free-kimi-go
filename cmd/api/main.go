package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/princetheprogrammerbtw/kimi-free-api-go/internal/chat"
	"github.com/princetheprogrammerbtw/kimi-free-api-go/internal/models"
	"github.com/princetheprogrammerbtw/kimi-free-api-go/internal/token"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)

	kimiClient := chat.NewKimiClient()
	tokenManager := token.NewTokenManager(kimiClient)
	chatHandler := chat.NewChatHandler(tokenManager, kimiClient)

	// BEAST MODE: Disable all timeouts for long-running streaming
	app := fiber.New(fiber.Config{
		AppName:      "Kimi Free API (Go Version)",
		ReadTimeout:  0,
		WriteTimeout: 0,
		IdleTimeout:  0,
	})

	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New())

	app.Get("/ping", func(c *fiber.Ctx) error { return c.SendString("pong") })
	v1 := app.Group("/v1")
	v1.Get("/models", models.HandleModels)
	v1.Post("/chat/completions", chatHandler.HandleCompletions)

	port := os.Getenv("PORT")
	if port == "" { port = "8788" }
	logrus.Infof("Service starting on port %s...", port)
	if err := app.Listen(":" + port); err != nil {
		logrus.Fatalf("Failed to start server: %v", err)
	}
}
