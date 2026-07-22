package routes

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"sistema-editorial/editora/backend/src/auth/entity"
)

type authService interface {
	Login(ctx context.Context, input entity.LoginRequest) (entity.LoginResponse, error)
	Me(ctx context.Context, userID string) (entity.CurrentUserResponse, error)
	ChangePassword(ctx context.Context, userID string, input entity.ChangePasswordRequest) error
}

func Register(app *fiber.App, service authService, requireAuth fiber.Handler) {
	group := app.Group("/api/auth")

	group.Post("/login", func(c *fiber.Ctx) error {
		var request entity.LoginRequest
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "payload invalido para login",
			})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		response, err := service.Login(ctx, request)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		return c.JSON(response)
	})

	group.Get("/me", requireAuth, func(c *fiber.Ctx) error {
		userID := strings.TrimSpace(c.Locals("auth_user_id").(string))
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		response, err := service.Me(ctx, userID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "usuario nao encontrado",
			})
		}

		return c.JSON(response)
	})

	group.Post("/change-password", requireAuth, func(c *fiber.Ctx) error {
		var request entity.ChangePasswordRequest
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "payload invalido para troca de senha",
			})
		}

		userID := strings.TrimSpace(c.Locals("auth_user_id").(string))
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := service.ChangePassword(ctx, userID, request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message": "senha atualizada com sucesso",
		})
	})
}
