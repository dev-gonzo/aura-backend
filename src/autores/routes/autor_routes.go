package routes

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"sistema-editorial/editora/backend/src/autores/entity"
	autorservice "sistema-editorial/editora/backend/src/autores/service"
)

type service interface {
	List(ctx context.Context, query entity.ListQuery) ([]entity.ListItem, error)
	FindByID(ctx context.Context, id string) (entity.DetailResponse, error)
	Create(ctx context.Context, request entity.CreateRequest) (string, error)
	Update(ctx context.Context, id string, request entity.UpdateRequest) error
}

func Register(
	app *fiber.App,
	svc service,
	requireAuth fiber.Handler,
	adminOnly fiber.Handler,
) {
	group := app.Group("/api/autores", requireAuth, adminOnly)

	group.Get("/", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		items, err := svc.List(ctx, entity.ListQuery{
			Search: c.Query("search"),
			Status: c.Query("status"),
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
		}
		return c.JSON(fiber.Map{"items": items})
	})

	group.Get("/:id", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		item, err := svc.FindByID(ctx, c.Params("id"))
		if err != nil {
			status := fiber.StatusInternalServerError
			if autorservice.IsValidationError(err) {
				status = fiber.StatusBadRequest
			}
			return c.Status(status).JSON(fiber.Map{"message": err.Error()})
		}
		return c.JSON(item)
	})

	group.Post("/", func(c *fiber.Ctx) error {
		var request entity.CreateRequest
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "payload invalido para autor"})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		id, err := svc.Create(ctx, request)
		if err != nil {
			status := fiber.StatusInternalServerError
			if autorservice.IsValidationError(err) {
				status = fiber.StatusBadRequest
			}
			return c.Status(status).JSON(fiber.Map{"message": err.Error()})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
	})

	group.Put("/:id", func(c *fiber.Ctx) error {
		var request entity.UpdateRequest
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "payload invalido para autor"})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := svc.Update(ctx, c.Params("id"), request)
		if err != nil {
			status := fiber.StatusInternalServerError
			if autorservice.IsValidationError(err) {
				status = fiber.StatusBadRequest
			}
			return c.Status(status).JSON(fiber.Map{"message": err.Error()})
		}

		return c.SendStatus(fiber.StatusNoContent)
	})
}
