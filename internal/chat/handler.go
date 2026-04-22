package chat

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/princetheprogrammerbtw/kimi-free-api-go/internal/core"
	"github.com/princetheprogrammerbtw/kimi-free-api-go/internal/token"
	"github.com/sirupsen/logrus"
)

type ChatHandler struct {
	tokenManager *token.TokenManager
	kimiClient   *KimiClient
}

func NewChatHandler(tokenManager *token.TokenManager, kimiClient *KimiClient) *ChatHandler {
	return &ChatHandler{
		tokenManager: tokenManager,
		kimiClient:   kimiClient,
	}
}

func (h *ChatHandler) HandleCompletions(c *fiber.Ctx) error {
	var req ChatCompletionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Authorization header required"})
	}

	var lastErr error
	for i := 0; i < 3; i++ {
		tokenInfo, err := h.tokenManager.GetToken(authHeader)
		if err != nil {
			lastErr = err
			time.Sleep(2 * time.Second)
			continue
		}

		convId := req.ConversationId
		if convId == "" {
			convId, err = h.kimiClient.CreateConversation(req.Model, "New Conversation", tokenInfo.AccessToken, tokenInfo.UserId)
			if err != nil {
				lastErr = err
				time.Sleep(2 * time.Second)
				continue
			}
		}

		if req.Stream {
			return h.handleStream(c, req, tokenInfo, convId)
		} else {
			return h.handleNonStream(c, req, tokenInfo, convId)
		}
	}

	logrus.Errorf("Failed after 3 retries: %v", lastErr)
	return c.Status(500).JSON(fiber.Map{"error": "failed after retries: " + lastErr.Error()})
}

func (h *ChatHandler) handleStream(c *fiber.Ctx, req ChatCompletionRequest, tokenInfo *core.TokenInfo, convId string) error {
	// GOD-MODE HEADERS: Prevent any client/proxy from closing the connection
	c.Set("Content-Type", "text/event-stream; charset=utf-8")
	c.Set("Cache-Control", "no-cache, no-transform")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no") // Disable Nginx buffering

