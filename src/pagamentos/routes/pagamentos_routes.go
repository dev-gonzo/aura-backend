package routes

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"sistema-editorial/editora/backend/src/pagamentos/entity"
	pagamentosservice "sistema-editorial/editora/backend/src/pagamentos/service"
)

type service interface {
	GetConfig(ctx context.Context) (entity.SettingsResponse, error)
	UpdateConfig(ctx context.Context, request entity.UpdateSettingsRequest) error
	CreateCheckout(ctx context.Context, request entity.CreateCheckoutRequest) (entity.CheckoutResponse, error)
}

func Register(app *fiber.App, svc service, requireAuth fiber.Handler, adminOnly fiber.Handler) {
	group := app.Group("/api/pagamentos", requireAuth, adminOnly)

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
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "payload invalido para configuracao de pagamentos"})
		}

		if err := svc.UpdateConfig(c.UserContext(), request); err != nil {
			status := fiber.StatusInternalServerError
			if pagamentosservice.IsValidationError(err) {
				status = fiber.StatusBadRequest
			}
			return c.Status(status).JSON(fiber.Map{"message": err.Error()})
		}

		return c.SendStatus(fiber.StatusNoContent)
	})

	group.Post("/checkout", func(c *fiber.Ctx) error {
		var request entity.CreateCheckoutRequest
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "payload invalido para checkout de pagamento"})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		response, err := svc.CreateCheckout(ctx, request)
		if err != nil {
			status := fiber.StatusInternalServerError
			if pagamentosservice.IsValidationError(err) {
				status = fiber.StatusBadRequest
			}
			return c.Status(status).JSON(fiber.Map{"message": err.Error()})
		}

		return c.JSON(response)
	})
}
