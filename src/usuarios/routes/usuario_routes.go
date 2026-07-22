package routes

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"

	"sistema-editorial/editora/backend/src/usuarios/entity"
	userservice "sistema-editorial/editora/backend/src/usuarios/service"
)

type createService interface {
	Create(ctx context.Context, input entity.CreateRequest) (entity.CreateResponse, error)
	Update(ctx context.Context, id string, input entity.UpdateRequest) (entity.CreateResponse, error)
	List(ctx context.Context, query entity.ListQuery) (entity.ListResponse, error)
	FindByID(ctx context.Context, id string) (entity.ListItem, error)
	Block(ctx context.Context, id string) (entity.StatusActionResponse, error)
	Activate(ctx context.Context, id string) (entity.StatusActionResponse, error)
	ResetPassword(ctx context.Context, id string) (entity.ResetPasswordResponse, error)
}

func Register(
	app *fiber.App,
	service createService,
	requireAuth fiber.Handler,
	adminOnly fiber.Handler,
	adminOrSelf fiber.Handler,
) {
	group := app.Group("/api/usuarios", requireAuth)

	group.Get("", adminOnly, func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		response, err := service.List(ctx, entity.ListQuery{
			Search:   c.Query("q"),
			Role:     c.Query("role"),
			Page:     c.QueryInt("page", 1),
			PageSize: c.QueryInt("page_size", 20),
		})
		if err != nil {
			if userservice.IsValidationError(err) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"message": err.Error(),
				})
			}

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		return c.JSON(response)
	})

	group.Get("/:id", adminOrSelf, func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		response, err := service.FindByID(ctx, c.Params("id"))
		if err != nil {
			if userservice.IsValidationError(err) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"message": err.Error(),
				})
			}

			if errors.Is(err, pgx.ErrNoRows) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"message": "usuario nao encontrado",
				})
			}

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		return c.JSON(response)
	})

	group.Post("", adminOnly, func(c *fiber.Ctx) error {
		var request entity.CreateRequest
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "payload invalido para cadastro de usuario",
			})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		response, err := service.Create(ctx, request)
		if err != nil {
			if userservice.IsValidationError(err) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"message": err.Error(),
				})
			}

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		return c.Status(fiber.StatusCreated).JSON(response)
	})

	group.Put("/:id", adminOrSelf, func(c *fiber.Ctx) error {
		var request entity.UpdateRequest
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "payload invalido para atualizacao de usuario",
			})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		currentRoles, _ := c.Locals("auth_roles").([]string)
		isAdmin := containsRole(currentRoles, "ADMIN")
		if !isAdmin {
			currentUser, err := service.FindByID(ctx, c.Params("id"))
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"message": "usuario nao encontrado",
					})
				}

				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"message": err.Error(),
				})
			}

			request.Papeis = currentUser.Papeis
		}

		response, err := service.Update(ctx, c.Params("id"), request)
		if err != nil {
			if userservice.IsValidationError(err) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"message": err.Error(),
				})
			}

			if errors.Is(err, pgx.ErrNoRows) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"message": "usuario nao encontrado",
				})
			}

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		return c.JSON(response)
	})

	group.Patch("/:id/block", adminOnly, func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		response, err := service.Block(ctx, c.Params("id"))
		if err != nil {
			if userservice.IsValidationError(err) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"message": err.Error(),
				})
			}

			if errors.Is(err, pgx.ErrNoRows) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"message": "usuario nao encontrado",
				})
			}

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		return c.JSON(response)
	})

	group.Patch("/:id/activate", adminOnly, func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		response, err := service.Activate(ctx, c.Params("id"))
		if err != nil {
			if userservice.IsValidationError(err) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"message": err.Error(),
				})
			}

			if errors.Is(err, pgx.ErrNoRows) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"message": "usuario nao encontrado",
				})
			}

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		return c.JSON(response)
	})

	group.Post("/:id/reset-password", adminOnly, func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		response, err := service.ResetPassword(ctx, c.Params("id"))
		if err != nil {
			if userservice.IsValidationError(err) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"message": err.Error(),
				})
			}

			if errors.Is(err, pgx.ErrNoRows) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"message": "usuario nao encontrado",
				})
			}

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		return c.JSON(response)
	})
}

func containsRole(roles []string, target string) bool {
	target = strings.TrimSpace(strings.ToUpper(target))
	for _, role := range roles {
		if strings.TrimSpace(strings.ToUpper(role)) == target {
			return true
		}
	}

	return false
}
