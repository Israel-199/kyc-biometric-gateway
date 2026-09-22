package response

import (
	"github.com/gofiber/fiber/v2"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

func JSON(c *fiber.Ctx, status int, success bool, message string, data interface{}, err interface{}, meta interface{}) error {
	return c.Status(status).JSON(Response{
		Success: success,
		Message: message,
		Data:    data,
		Error:   err,
		Meta:    meta,
	})
}

func Success(c *fiber.Ctx, status int, message string, data interface{}) error {
	return JSON(c, status, true, message, data, nil, nil)
}

func SuccessWithMeta(c *fiber.Ctx, status int, message string, data interface{}, meta interface{}) error {
	return JSON(c, status, true, message, data, nil, meta)
}

func Error(c *fiber.Ctx, status int, message string, err interface{}) error {
	return JSON(c, status, false, message, nil, err, nil)
}
