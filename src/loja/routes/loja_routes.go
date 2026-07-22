package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"sistema-editorial/editora/backend/src/loja/entity"
	lojaservice "sistema-editorial/editora/backend/src/loja/service"
)

type service interface {
	GetSettings(ctx context.Context) (entity.SettingsResponse, error)
	UpdateSettings(ctx context.Context, request entity.UpdateSettingsRequest) error
	UpdateIntegrations(ctx context.Context, settings entity.IntegrationSettings) error
	PublishSettings(ctx context.Context) error
	ListCategories(ctx context.Context) ([]entity.AdminStoreCategoryListItem, error)
	CreateCategory(ctx context.Context, request entity.StoreCategoryPayload) (string, error)
	UpdateCategory(ctx context.Context, id string, request entity.StoreCategoryPayload) error
	DeleteCategory(ctx context.Context, id string) error
	ListAdminProducts(ctx context.Context) ([]entity.AdminProductListItem, error)
	GetAdminProductByID(ctx context.Context, id string) (entity.AdminProductDetail, error)
	CreateProduct(ctx context.Context, request entity.ProductPayload) (string, error)
	UpdateProduct(ctx context.Context, id string, request entity.ProductPayload) error
	GetPublicConfig(ctx context.Context, preview bool) (entity.PublicConfigResponse, error)
	ListPublicProducts(ctx context.Context) ([]entity.PublicProductListItem, error)
	GetPublicProductBySlug(ctx context.Context, slug string) (entity.PublicProductDetail, error)
}

func Register(app *fiber.App, svc service, requireAuth fiber.Handler, adminOnly fiber.Handler) {
	adminGroup := app.Group("/api/loja", requireAuth, adminOnly)

	adminGroup.Get("/configuracao", func(c *fiber.Ctx) error {
		response, err := svc.GetSettings(c.UserContext())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
		}
		c.Set("X-Debug-Menu-Background-Mode", response.MenuBackgroundMode)
		// #region debug-point C:admin-get-config-response
		go http.Post("http://127.0.0.1:7777/event", "application/json", bytes.NewBufferString(string(func() []byte { payload, _ := json.Marshal(fiber.Map{"sessionId": "menu-bg-persist", "runId": "pre-fix", "hypothesisId": "C", "location": "loja_routes.go:GET /api/loja/configuracao", "msg": "[DEBUG] backend GET /loja/configuracao response snapshot", "data": fiber.Map{"menu_background_mode": response.MenuBackgroundMode, "content_width_mode": response.ContentWidthMode, "draft_updated_at": response.DraftUpdatedAt, "updated_at": response.UpdatedAt}, "ts": time.Now().UnixMilli()}); return payload }())))
		// #endregion
		return c.JSON(response)
	})

	adminGroup.Put("/configuracao", func(c *fiber.Ctx) error {
		var request entity.UpdateSettingsRequest
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "payload invalido para configuracao da loja"})
		}
		c.Set("X-Debug-Request-Menu-Background-Mode", request.MenuBackgroundMode)
		// #region debug-point B:admin-put-config-request
		go http.Post("http://127.0.0.1:7777/event", "application/json", bytes.NewBufferString(string(func() []byte { payload, _ := json.Marshal(fiber.Map{"sessionId": "menu-bg-persist", "runId": "pre-fix", "hypothesisId": "B", "location": "loja_routes.go:PUT /api/loja/configuracao", "msg": "[DEBUG] backend received payload in PUT /loja/configuracao", "data": fiber.Map{"menu_background_mode": request.MenuBackgroundMode, "content_width_mode": request.ContentWidthMode, "menu_color": request.MenuColor, "menu_text_color": request.MenuTextColor}, "ts": time.Now().UnixMilli()}); return payload }())))
		// #endregion
		if err := svc.UpdateSettings(c.UserContext(), request); err != nil {
			status := fiber.StatusInternalServerError
			if lojaservice.IsValidationError(err) {
				status = fiber.StatusBadRequest
			}
			return c.Status(status).JSON(fiber.Map{"message": err.Error()})
		}
		afterSave, afterErr := svc.GetSettings(c.UserContext())
		if afterErr == nil {
			c.Set("X-Debug-After-Save-Menu-Background-Mode", afterSave.MenuBackgroundMode)
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	adminGroup.Put("/configuracao/integracoes", func(c *fiber.Ctx) error {
		var request entity.IntegrationSettings
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "payload invalido para integracoes da loja"})
		}
		if err := svc.UpdateIntegrations(c.UserContext(), request); err != nil {
			status := fiber.StatusInternalServerError
			if lojaservice.IsValidationError(err) {
				status = fiber.StatusBadRequest
			}
			return c.Status(status).JSON(fiber.Map{"message": err.Error()})
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	adminGroup.Post("/configuracao/publicar", func(c *fiber.Ctx) error {
		if err := svc.PublishSettings(c.UserContext()); err != nil {
			status := fiber.StatusInternalServerError
			if lojaservice.IsValidationError(err) {
				status = fiber.StatusBadRequest
			}
			return c.Status(status).JSON(fiber.Map{"message": err.Error()})
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	adminGroup.Get("/categorias", func(c *fiber.Ctx) error {
		response, err := svc.ListCategories(c.UserContext())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
		}
		return c.JSON(fiber.Map{"items": response})
	})

	adminGroup.Post("/categorias", func(c *fiber.Ctx) error {
		var request entity.StoreCategoryPayload
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "payload invalido para categoria da loja"})
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		id, err := svc.CreateCategory(ctx, request)
		if err != nil {
			status := fiber.StatusInternalServerError
			if lojaservice.IsValidationError(err) {
				status = fiber.StatusBadRequest
			}
			return c.Status(status).JSON(fiber.Map{"message": err.Error()})
		}
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
	})

	adminGroup.Put("/categorias/:id", func(c *fiber.Ctx) error {
		var request entity.StoreCategoryPayload
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "payload invalido para categoria da loja"})
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := svc.UpdateCategory(ctx, c.Params("id"), request); err != nil {
			status := fiber.StatusInternalServerError
			if lojaservice.IsValidationError(err) {
				status = fiber.StatusBadRequest
			}
			return c.Status(status).JSON(fiber.Map{"message": err.Error()})
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	adminGroup.Delete("/categorias/:id", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := svc.DeleteCategory(ctx, c.Params("id")); err != nil {
			status := fiber.StatusInternalServerError
			if lojaservice.IsValidationError(err) {
				status = fiber.StatusBadRequest
			}
			return c.Status(status).JSON(fiber.Map{"message": err.Error()})
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	adminGroup.Get("/produtos", func(c *fiber.Ctx) error {
		response, err := svc.ListAdminProducts(c.UserContext())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
		}
		return c.JSON(fiber.Map{"items": response})
	})

	adminGroup.Get("/produtos/:id", func(c *fiber.Ctx) error {
		response, err := svc.GetAdminProductByID(c.UserContext(), c.Params("id"))
		if err != nil {
			status := fiber.StatusInternalServerError
			if lojaservice.IsValidationError(err) {
				status = fiber.StatusNotFound
			}
			return c.Status(status).JSON(fiber.Map{"message": err.Error()})
		}
		return c.JSON(response)
	})

	adminGroup.Post("/produtos", func(c *fiber.Ctx) error {
		var request entity.ProductPayload
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "payload invalido para produto da loja"})
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		id, err := svc.CreateProduct(ctx, request)
		if err != nil {
			status := fiber.StatusInternalServerError
			if lojaservice.IsValidationError(err) {
				status = fiber.StatusBadRequest
			}
			return c.Status(status).JSON(fiber.Map{"message": err.Error()})
		}
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
	})

	adminGroup.Put("/produtos/:id", func(c *fiber.Ctx) error {
		var request entity.ProductPayload
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "payload invalido para produto da loja"})
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := svc.UpdateProduct(ctx, c.Params("id"), request); err != nil {
			status := fiber.StatusInternalServerError
			if lojaservice.IsValidationError(err) {
				status = fiber.StatusBadRequest
			}
			return c.Status(status).JSON(fiber.Map{"message": err.Error()})
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	publicGroup := app.Group("/api/publico/loja")

	publicGroup.Get("/configuracao", func(c *fiber.Ctx) error {
		preview := c.Query("preview") == "draft"
		response, err := svc.GetPublicConfig(c.UserContext(), preview)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
		}
		return c.JSON(response)
	})

	publicGroup.Get("/produtos", func(c *fiber.Ctx) error {
		response, err := svc.ListPublicProducts(c.UserContext())
		if err != nil {
			// #region debug-point A:public-products-error
			go http.Post("http://127.0.0.1:7777/event", "application/json", bytes.NewBufferString(string(func() []byte { payload, _ := json.Marshal(fiber.Map{"sessionId": "store-produtos-500", "runId": "pre-fix", "hypothesisId": "A", "location": "loja_routes.go:GET /api/publico/loja/produtos", "msg": "[DEBUG] backend failed while listing public store products", "data": fiber.Map{"error": err.Error()}, "ts": time.Now().UnixMilli()}); return payload }())))
			// #endregion
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
		}
		return c.JSON(fiber.Map{"items": response})
	})

	publicGroup.Get("/produtos/:slug", func(c *fiber.Ctx) error {
		response, err := svc.GetPublicProductBySlug(c.UserContext(), c.Params("slug"))
		if err != nil {
			status := fiber.StatusInternalServerError
			if lojaservice.IsValidationError(err) {
				status = fiber.StatusNotFound
			}
			return c.Status(status).JSON(fiber.Map{"message": err.Error()})
		}
		return c.JSON(response)
	})
}
