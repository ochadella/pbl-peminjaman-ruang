package route

import (
	"github.com/gin-gonic/gin"
	"github.com/ochadella/pbl-peminjaman-ruang/app/repository"
	"github.com/ochadella/pbl-peminjaman-ruang/app/service"
	"gorm.io/gorm"
)

func RegisterAuthRoutes(api *gin.RouterGroup, db *gorm.DB) {
	repo := repository.NewUserRepository(db)
	authH := NewAuthHandler(service.NewAuthService(repo))

	api.POST("/auth/register", authH.Register)
	api.POST("/auth/login", authH.Login)
}