package utils

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

type Utils interface {
	ErrorWrapper(err error) error
}

type utils struct{}

func NewUtils() Utils {
	return &utils{}
}

// ErrorWrapper will wrap the error with a fiber error when it's an unhandled error object
func (u *utils) ErrorWrapper(err error) error {
	if we, ok := err.(mongo.WriteException); ok {
		for _, wce := range we.WriteErrors {
			if wce.Code == 11000 || wce.Code == 11001 || wce.Code == 12582 || wce.Code == 16460 && strings.Contains(wce.Message, " E11000 ") {
				return fiber.NewError(fiber.StatusConflict, "Model Already Exists")
			}
		}
	}
	if err.Error() == "mongo: no documents in result" {
		return fiber.NewError(fiber.StatusOK, "No Operation Executed")
	}
	return err
}
