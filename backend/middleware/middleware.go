package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"resumetake/models"
)

// ExtractBearerToken returns the token from an "Authorization: Bearer <token>"
// header, accepting any case of the scheme name (Bearer/bearer/BEARER) per
// RFC 6750. Returns "" if the header is missing or malformed.
func ExtractBearerToken(c *fiber.Ctx) string {
	h := c.Get("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return h[len("Bearer "):]
	}
	if len(h) > 7 && strings.EqualFold(h[:7], "Bearer ") {
		return h[7:]
	}
	return ""
}

func AuthMiddleware(userStore *models.UserStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{
				"error":   "UNAUTHORIZED",
				"message": "Missing authorization header",
			})
		}

		token := ExtractBearerToken(c)
		if token == "" {
			return c.Status(401).JSON(fiber.Map{
				"error":   "UNAUTHORIZED",
				"message": "Invalid authorization format",
			})
		}
		// R58-B-L4: reject abnormally long tokens before hashing. Valid tokens
		// are 64-char hex (32 bytes). An attacker could send multi-KB tokens
		// to waste CPU on map hash computation; the length check is a cheap
		// O(1) filter.
		if len(token) > 128 {
			return c.Status(401).JSON(fiber.Map{
				"error":   "UNAUTHORIZED",
				"message": "Invalid token",
			})
		}

		user, ok := userStore.GetByToken(token)
		if !ok {
			return c.Status(401).JSON(fiber.Map{
				"error":   "UNAUTHORIZED",
				"message": "Invalid token",
			})
		}

		c.Locals("user", user)
		return c.Next()
	}
}
