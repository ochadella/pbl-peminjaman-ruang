package route

import (
	"context"
	"time"

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
	violationRepo := repository.NewViolationRepository(db)
	bookingRepo := repository.NewBookingRepository(db)

	authService := service.NewAuthService(userRepo, jwtSecret)
	userService := service.NewUserService(userRepo)
	roomService := service.NewRoomService(roomRepo)
	hourService := service.NewOperatingHourService(hourRepo)
	violationService := service.NewViolationService(violationRepo, userRepo)
	attendanceService := service.NewAttendanceService(bookingRepo, roomRepo, hourRepo, violationRepo)
	bookingService := service.NewBookingService(bookingRepo, roomRepo, hourRepo, userRepo)

	// Cek no-show otomatis setiap 1 menit di background.
	attendanceService.StartNoShowWorker(context.Background(), time.Minute)

	auth := middleware.Auth(jwtSecret)
	adminOnly := middleware.RequireRole(model.RoleAdmin)
	staffOnly := middleware.RequireRole(model.RolePetugas, model.RoleAdmin)
	studentOnly := middleware.RequireRole(model.RoleMahasiswa)

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
	rooms.Get("/:id/availability", attendanceService.Availability)
	rooms.Post("/", adminOnly, roomService.Create)
	rooms.Patch("/:id", adminOnly, roomService.Update)

	hours := api.Group("/operating-hours")
	hours.Get("/", hourService.List)
	hours.Put("/:day", auth, adminOnly, hourService.Update)

	violations := api.Group("/violations", auth)
	violations.Get("/me", violationService.ListMine)
	violations.Get("/user/:id", staffOnly, violationService.ListByUser)
	violations.Post("/", staffOnly, violationService.Create)

	attendance := api.Group("/attendance", auth)
	attendance.Post("/check-in", attendanceService.CheckIn)
	attendance.Post("/process-no-show", staffOnly, attendanceService.RunNoShow)
	attendance.Post("/:id/check-out", attendanceService.CheckOut)

	bookings := api.Group("/bookings", auth)
	bookings.Get("/me", bookingService.ListMine) 
	bookings.Get("/:id", bookingService.Get)
	bookings.Post("/", studentOnly, bookingService.Create)
	bookings.Patch("/:id/cancel", studentOnly, bookingService.Cancel)
}
