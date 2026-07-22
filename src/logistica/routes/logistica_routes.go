package routes

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"sistema-editorial/editora/backend/src/logistica/entity"
	logisticaservice "sistema-editorial/editora/backend/src/logistica/service"
)

type service interface {
	GetConfig(ctx context.Context) (entity.SettingsResponse, error)
	UpdateConfig(ctx context.Context, request entity.UpdateSettingsRequest) error
	CalculateQuote(ctx context.Context, request entity.QuoteRequest) (entity.QuoteResponse, error)
}

func Register(app *fiber.App, svc service, requireAuth fiber.Handler, adminOnly fiber.Handler) {
	group := app.Group("/api/logistica", requireAuth, adminOnly)

	group.Get("/configuracao", func(c *fiber.Ctx) error {
		response, err := svc.GetConfig(c.UserContext())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
		}
		return c.JSON(response)
	})

	group.Put("/configuracao", func(c *fiber.Ctx) error {
		var request entity.UpdateSettingsRequest
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "payload invalido para configuracao de logistica"})
		}

		if err := svc.UpdateConfig(c.UserContext(), request); err != nil {
			status := fiber.StatusInternalServerError
			if logisticaservice.IsValidationError(err) {
				status = fiber.StatusBadRequest
			}
			return c.Status(status).JSON(fiber.Map{"message": err.Error()})
		}

		return c.SendStatus(fiber.StatusNoContent)
	})

	group.Post("/cotacoes", func(c *fiber.Ctx) error {
		var request entity.QuoteRequest
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "payload invalido para cotacao de logistica"})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		response, err := svc.CalculateQuote(ctx, request)
		if err != nil {
			status := fiber.StatusInternalServerError
			if logisticaservice.IsValidationError(err) {
				status = fiber.StatusBadRequest
			}
			return c.Status(status).JSON(fiber.Map{"message": err.Error()})
		}

		return c.JSON(response)
	})
}
