package routes

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"sistema-editorial/editora/backend/src/livros/entity"
	livroservice "sistema-editorial/editora/backend/src/livros/service"
)

type service interface {
	List(ctx context.Context, query entity.ListQuery) ([]entity.ListItem, error)
	FindByID(ctx context.Context, id string) (entity.DetailResponse, error)
	Create(ctx context.Context, request entity.CreateRequest) (string, error)
	Update(ctx context.Context, id string, request entity.UpdateRequest) error
	RegisterStockMovement(ctx context.Context, id string, request entity.StockMovementRequest) error
	Validate(ctx context.Context, input entity.ValidationRequest) (entity.ValidationResponse, error)
}

func Register(
	app *fiber.App,
	service service,
	requireAuth fiber.Handler,
	adminOnly fiber.Handler,
) {
	group := app.Group("/api/livros", requireAuth, adminOnly)

	group.Get("/", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		items, err := service.List(ctx, entity.ListQuery{
			Search:  c.Query("search"),
			Status:  c.Query("status"),
			AutorID: c.Query("autor_id"),
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
		}

		return c.JSON(fiber.Map{"items": items})
	})

	group.Get("/:id", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		response, err := service.FindByID(ctx, c.Params("id"))
		if err != nil {
			status := fiber.StatusInternalServerError
			if livroservice.IsValidationError(err) {
				status = fiber.StatusBadRequest
			}
			return c.Status(status).JSON(fiber.Map{"message": err.Error()})
		}

		return c.JSON(response)
	})

	group.Post("/", func(c *fiber.Ctx) error {
		var request entity.CreateRequest
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "payload invalido para livro"})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		id, err := service.Create(ctx, request)
		if err != nil {
			status := fiber.StatusInternalServerError
			if livroservice.IsValidationError(err) {
				status = fiber.StatusBadRequest
			}
			return c.Status(status).JSON(fiber.Map{"message": err.Error()})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
	})

	group.Put("/:id", func(c *fiber.Ctx) error {
		var request entity.UpdateRequest
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "payload invalido para livro"})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := service.Update(ctx, c.Params("id"), request); err != nil {
			status := fiber.StatusInternalServerError
			if livroservice.IsValidationError(err) {
				status = fiber.StatusBadRequest
			}
			return c.Status(status).JSON(fiber.Map{"message": err.Error()})
		}

		return c.SendStatus(fiber.StatusNoContent)
	})

	group.Post("/:id/estoque/movimentos", func(c *fiber.Ctx) error {
		var request entity.StockMovementRequest
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "payload invalido para movimento de estoque"})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := service.RegisterStockMovement(ctx, c.Params("id"), request); err != nil {
			status := fiber.StatusInternalServerError
			if livroservice.IsValidationError(err) {
				status = fiber.StatusBadRequest
			}
			return c.Status(status).JSON(fiber.Map{"message": err.Error()})
		}

		return c.SendStatus(fiber.StatusNoContent)
	})

	group.Post("/validate", func(c *fiber.Ctx) error {
		var request entity.ValidationRequest
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "payload invalido para validacao de livro",
			})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		response, err := service.Validate(ctx, request)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		return c.JSON(response)
	})
}
