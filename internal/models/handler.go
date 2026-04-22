package models

import (
	"github.com/gofiber/fiber/v2"
)

func HandleModels(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"data": []fiber.Map{
			{"id": "kimi-k2.6-thinking", "object": "model", "owned_by": "moonshot"},
			{"id": "kimi-k2.6-instant", "object": "model", "owned_by": "moonshot"},
			{"id": "kimi-k2.6-thinking-deep", "object": "model", "owned_by": "moonshot"},
			{"id": "kimi-k2.6-agent", "object": "model", "owned_by": "moonshot"},
			{"id": "kimi-k2.6-agent-swarm", "object": "model", "owned_by": "moonshot"},
		},
	})
}
