package route

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ochadella/pbl-peminjaman-ruang/app/service"
)

type AuthHandler struct{ svc service.AuthService }

func NewAuthHandler(svc service.AuthService) *AuthHandler { return &AuthHandler{svc} }

func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.svc.Register(req)
	if err != nil {
		code := http.StatusInternalServerError
		if errors.Is(err, service.ErrEmailTaken) || errors.Is(err, service.ErrNimNipTaken) {
			code = http.StatusConflict
		}
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "registrasi berhasil", "data": user})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, user, err := h.svc.Login(req)
	if err != nil {
		code := http.StatusInternalServerError
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			code = http.StatusUnauthorized
		case errors.Is(err, service.ErrAccountInactive):
			code = http.StatusForbidden
		}
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "data": user})
}