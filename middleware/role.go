package middleware

import (
	"github.com/gofiber/fiber/v2"

	"github.com/ochadella/pbl-peminjaman-ruang/helper"
)

// RequireRole hanya meloloskan user dengan salah satu role yang disebut.
// Harus dipasang SETELAH Auth. Contoh: RequireRole("admin", "petugas").
func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals(LocalRole).(string)
		for _, r := range roles {
			if r == role {
				return c.Next()
			}
		}
		return helper.ErrorResponse(c, fiber.StatusForbidden, "anda tidak punya akses ke fitur ini")
	}
}
