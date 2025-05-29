package middleware

import (
	"log"
	"strings"

	"github.com/Harshal292004/subscription-service/internal/utils"
	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		log.Printf("[AuthMiddleware] === Starting authentication for path: %s ===", c.Path())

		authHeader := c.Get("Authorization")
		log.Printf("[AuthMiddleware] Authorization header length: %d", len(authHeader))

		if authHeader == "" {
			log.Println("[AuthMiddleware] No Authorization header found")
			log.Println("[AuthMiddleware] === Returning 401 error ===")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing token"})
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			log.Printf("[AuthMiddleware] Authorization header doesn't start with 'Bearer ': %s", authHeader[:min(len(authHeader), 20)])
			log.Println("[AuthMiddleware] === Returning 401 error ===")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing token"})
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		log.Printf("[AuthMiddleware] Extracted token, length: %d", len(tokenStr))
		log.Printf("[AuthMiddleware] Token preview: %s...", tokenStr[:min(len(tokenStr), 20)])

		userID, err := utils.ValidateSession(tokenStr)
		if err != nil {
			log.Printf("[AuthMiddleware] Token validation failed: %v", err)
			log.Println("[AuthMiddleware] === Returning 401 error ===")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
		}

		log.Printf("[AuthMiddleware] Token validation successful for userID: %d", userID)

		// Store as int to match handler expectations
		userIDInt := int(userID)
		c.Locals("userId", userIDInt)
		log.Printf("[AuthMiddleware] Stored userID in context as int: %d", userIDInt)

		// Verify the stored value
		storedValue := c.Locals("userId")
		log.Printf("[AuthMiddleware] Verification - stored value: %v (type: %T)", storedValue, storedValue)

		log.Printf("[AuthMiddleware] === Authentication successful, proceeding to next handler ===")
		return c.Next()
	}
}
