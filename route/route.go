package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ochadella/pbl-peminjaman-ruang/app/model"
	"github.com/ochadella/pbl-peminjaman-ruang/app/repository"
	"github.com/ochadella/pbl-peminjaman-ruang/app/service"
	"github.com/ochadella/pbl-peminjaman-ruang/middleware"
)

func SetupRoutes(app *fiber.App, db *pgxpool.Pool, jwtSecret string) {
	userRepo := repository.NewUserRepository(db)
	roomRepo := repository.NewRoomRepository(db)
	hourRepo := repository.NewOperatingHourRepository(db)

	authService := service.NewAuthService(userRepo, jwtSecret)
	userService := service.NewUserService(userRepo)
	roomService := service.NewRoomService(roomRepo)
	hourService := service.NewOperatingHourService(hourRepo)

	auth := middleware.Auth(jwtSecret)
	adminOnly := middleware.RequireRole(model.RoleAdmin)

	api := app.Group("/api/v1")

	api.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "ok",
			"message": "RuangKu API is running",
		})
	})

	authGroup := api.Group("/auth")
	authGroup.Post("/register", authService.Register)
	authGroup.Post("/login", authService.Login)
	authGroup.Get("/me", auth, authService.Me)

	users := api.Group("/users", auth, adminOnly)
	users.Get("/", userService.List)
	users.Patch("/:id/role", userService.UpdateRole)

	rooms := api.Group("/rooms", auth)
	rooms.Get("/", roomService.List)
	rooms.Get("/:id", roomService.Get)
	rooms.Post("/", adminOnly, roomService.Create)
	rooms.Patch("/:id", adminOnly, roomService.Update)

	hours := api.Group("/operating-hours")
	hours.Get("/", hourService.List)
	hours.Put("/:day", auth, adminOnly, hourService.Update)
}
