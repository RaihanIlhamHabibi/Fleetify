package middleware

import (
	"strconv"

	"fleetify/models"
	"fleetify/repository"

	"github.com/gofiber/fiber/v2"
)

func RequireUser(userRepo *repository.UserRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userIDStr := c.Get("X-User-ID")
		if userIDStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "X-User-ID header is required",
			})
		}
		userID, err := strconv.ParseUint(userIDStr, 10, 64)
		if err != nil || userID == 0 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid X-User-ID",
			})
		}
		user, err := userRepo.FindByID(userID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "user not found",
			})
		}
		c.Locals("user", user)
		return c.Next()
	}
}

func RequireRole(roles ...models.UserRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := c.Locals("user").(*models.User)
		if !ok || user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}
		for _, role := range roles {
			if user.Role == role {
				return c.Next()
			}
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "insufficient role",
		})
	}
}
