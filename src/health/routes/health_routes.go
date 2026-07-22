package routes

import (
	"github.com/gofiber/fiber/v2"

	"sistema-editorial/editora/backend/src/health/entity"
)

type statusService interface {
	Status() entity.StatusResponse
}

func Register(app *fiber.App, service statusService) {
	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(service.Status())
	})
}
