package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/ochadella/pbl-peminjaman-ruang/helper"
)

const (
	LocalUserID = "user_id"
	LocalRole   = "role"
)

// Auth memastikan request membawa header "Authorization: Bearer <token>"
// yang valid, lalu menyimpan user_id dan role ke c.Locals.
func Auth(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenStr, ok := strings.CutPrefix(c.Get("Authorization"), "Bearer ")
		if !ok || tokenStr == "" {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "silakan login terlebih dahulu")
		}
		claims, err := helper.ParseToken(secret, tokenStr)
		if err != nil {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
		}
		c.Locals(LocalUserID, claims.UserID)
		c.Locals(LocalRole, claims.Role)
		return c.Next()
	}
}

// CurrentUserID mengambil id user yang sedang login (diisi oleh Auth).
func CurrentUserID(c *fiber.Ctx) int64 {
	id, _ := c.Locals(LocalUserID).(int64)
	return id
}
