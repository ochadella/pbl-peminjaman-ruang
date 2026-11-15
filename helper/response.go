package helper

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func SuccessResponse(c *fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusOK).JSON(Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func ErrorResponse(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(Response{
		Success: false,
		Message: message,
	})
}

func HandleError(c *fiber.Ctx, err error) error {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return ErrorResponse(c, appErr.StatusCode, appErr.Message)
	}

	return ErrorResponse(
		c,
		fiber.StatusInternalServerError,
		"Terjadi kesalahan pada server",
	)
}
