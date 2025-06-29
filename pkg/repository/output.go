package models

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Error struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type Output struct {
	Data  interface{} `json:"data"`
	Error []Error     `json:"error"`
}

func NewOutput(data interface{}, errors ...Error) map[string]interface{} {
	return map[string]interface{}{
		"data":  data,
		"error": errors,
	}
}

func NewError(message string, code int) Error {
	return Error{
		Message: message,
		Code:    code,
	}
}
func NewErrors(errors ...error) []Error {
	errs := []Error{}
	for _, err := range errors {
		errs = append(errs, Error{
			Message: err.Error(),
			Code:    fiber.StatusInternalServerError,
		})
	}
	return errs
}

func AbortTransactionAndReturnError(ctx context.Context, session *mongo.Session, c *fiber.Ctx, err error) error {

	return c.Status(fiber.StatusInternalServerError).JSON(NewOutput(nil, Error{
		Message: err.Error(),
		Code:    fiber.StatusInternalServerError,
	}))
}

func ReturnError(c *fiber.Ctx, err error) error {
	return c.Status(fiber.StatusInternalServerError).JSON(NewOutput(nil, Error{
		Message: err.Error(),
		Code:    fiber.StatusInternalServerError,
	}))
}
