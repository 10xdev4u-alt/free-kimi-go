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
