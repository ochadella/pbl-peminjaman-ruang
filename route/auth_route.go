package route

import (
	"github.com/gin-gonic/gin"
	"github.com/ochadella/pbl-peminjaman-ruang/app/repository"
	"github.com/ochadella/pbl-peminjaman-ruang/app/service"
	"github.com/ochadella/pbl-peminjaman-ruang/middleware"
	"gorm.io/gorm"
)

func RegisterAuthRoutes(api *gin.RouterGroup, db *gorm.DB) {
	repo := repository.NewUserRepository(db)
	authH := NewAuthHandler(service.NewAuthService(repo))

	api.POST("/auth/register", authH.Register)
	api.POST("/auth/login", authH.Login)
	me := api.Group("/users", middleware.AuthMiddleware())
	me.GET("/me", userH.Me)
	me.PUT("/me", userH.UpdateMe)

	admin := api.Group("/admin/users", middleware.AuthMiddleware(), middleware.RoleMiddleware("admin"))
	admin.GET("", userH.List)
	admin.PUT("/:id/role", userH.UpdateRole)
	admin.PUT("/:id/status", userH.UpdateStatus)
}