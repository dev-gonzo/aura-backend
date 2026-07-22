package routes

import (
	"slices"
	"strings"

	"github.com/gofiber/fiber/v2"

	authservice "sistema-editorial/editora/backend/src/auth/service"
)

type tokenParser interface {
	ParseToken(tokenString string) (*authservice.Claims, error)
}

func RequireAuth(service tokenParser) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := strings.TrimSpace(c.Get("Authorization"))
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "token de acesso nao informado",
			})
		}

		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		claims, err := service.ParseToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "token invalido",
			})
		}

		c.Locals("auth_user_id", claims.Subject)
		c.Locals("auth_roles", claims.Papeis)
		c.Locals("auth_must_change_password", claims.PrecisaTrocarSenha)

		return c.Next()
	}
}

func RequireRoles(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		currentRoles, ok := c.Locals("auth_roles").([]string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "usuario sem papeis validos",
			})
		}

		for _, role := range roles {
			if slices.Contains(currentRoles, role) {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "usuario sem permissao para esta operacao",
		})
	}
}

func RequireRolesOrSelfParam(paramName string, roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		currentUserID, _ := c.Locals("auth_user_id").(string)
		targetUserID := strings.TrimSpace(c.Params(paramName))

		if currentUserID != "" && targetUserID != "" && currentUserID == targetUserID {
			return c.Next()
		}

		return RequireRoles(roles...)(c)
	}
}
