package routes

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	editalentity "sistema-editorial/editora/backend/src/editais/entity"
	editalservice "sistema-editorial/editora/backend/src/editais/service"
)

type uploadService interface {
	Upload(ctx context.Context, input editalentity.UploadRequest) (editalentity.UploadResponse, error)
	List(ctx context.Context, query editalentity.ListQuery) ([]editalentity.ListItem, error)
	FindByID(ctx context.Context, id string) (editalentity.DetailResponse, error)
	Create(ctx context.Context, input editalentity.CreateRequest) (editalentity.DetailResponse, error)
	Update(ctx context.Context, id string, input editalentity.UpdateRequest) (editalentity.DetailResponse, error)
}

func Register(app *fiber.App, service uploadService, requireAuth fiber.Handler, adminOnly fiber.Handler) {
	group := app.Group("/api/editais")
	group.Use(requireAuth, adminOnly)

	group.Get("/", func(c *fiber.Ctx) error {
		response, err := service.List(context.Background(), editalentity.ListQuery{
			Search: c.Query("search"),
			Status: c.Query("status"),
		})
		if err != nil {
			statusCode := fiber.StatusInternalServerError
			if editalservice.IsValidationError(err) {
				statusCode = fiber.StatusBadRequest
			}

			return c.Status(statusCode).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"items": response,
		})
	})

	group.Get("/:id", func(c *fiber.Ctx) error {
		response, err := service.FindByID(context.Background(), c.Params("id"))
		if err != nil {
			statusCode := fiber.StatusInternalServerError
			if editalservice.IsValidationError(err) {
				statusCode = fiber.StatusBadRequest
			}

			return c.Status(statusCode).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		return c.JSON(response)
	})

	group.Post("/", func(c *fiber.Ctx) error {
		var payload editalentity.CreateRequest
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "nao foi possivel ler os dados do edital",
			})
		}

		response, err := service.Create(context.Background(), payload)
		if err != nil {
			statusCode := fiber.StatusInternalServerError
			if editalservice.IsValidationError(err) {
				statusCode = fiber.StatusBadRequest
			}

			return c.Status(statusCode).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		return c.Status(fiber.StatusCreated).JSON(response)
	})

	group.Put("/:id", func(c *fiber.Ctx) error {
		var payload editalentity.UpdateRequest
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "nao foi possivel ler os dados do edital",
			})
		}

		response, err := service.Update(context.Background(), c.Params("id"), payload)
		if err != nil {
			statusCode := fiber.StatusInternalServerError
			if editalservice.IsValidationError(err) {
				statusCode = fiber.StatusBadRequest
			}

			return c.Status(statusCode).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		return c.JSON(response)
	})

	group.Post("/upload", func(c *fiber.Ctx) error {
		fileHeader, err := c.FormFile("arquivo")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "selecione um arquivo",
			})
		}

		file, err := fileHeader.Open()
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "nao foi possivel ler o arquivo",
			})
		}
		defer file.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		response, err := service.Upload(ctx, editalentity.UploadRequest{
			FileName:    fileHeader.Filename,
			ContentType: fileHeader.Header.Get("Content-Type"),
			Size:        fileHeader.Size,
			Body:        file,
		})
		if err != nil {
			statusCode := fiber.StatusInternalServerError
			if editalservice.IsValidationError(err) {
				statusCode = fiber.StatusBadRequest
			}

			return c.Status(statusCode).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		return c.JSON(response)
	})
}
