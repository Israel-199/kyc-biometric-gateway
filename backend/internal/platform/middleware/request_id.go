package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const HeaderXCorrelationID = "X-Correlation-ID"

func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		correlationID := c.Get(HeaderXCorrelationID)
		if correlationID == "" {
			correlationID = uuid.New().String()
		}
		c.Set(HeaderXCorrelationID, correlationID)
		c.Locals("correlation_id", correlationID)
		return c.Next()
	}
}
