package middleware

import (
	"strings"

	"github.com/Harshal292004/subscription-service/internal/repository"
	"github.com/Harshal292004/subscription-service/internal/utils"
	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(repo *repository.Repository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing token"})
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		userID, err := utils.ValidateSession(tokenStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
		}
		c.Locals("user_id", userID)
		return c.Next()
	}
}
