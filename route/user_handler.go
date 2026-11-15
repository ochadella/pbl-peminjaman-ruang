package route

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ochadella/pbl-peminjaman-ruang/app/service"
	"gorm.io/gorm"
)

type UserHandler struct{ svc service.UserService }

func NewUserHandler(svc service.UserService) *UserHandler { return &UserHandler{svc} }

func fail(c *gin.Context, err error) {
	code := http.StatusInternalServerError
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		code = http.StatusNotFound
	case errors.Is(err, service.ErrInvalidRole):
		code = http.StatusBadRequest
	}
	c.JSON(code, gin.H{"error": err.Error()})
}

func (h *UserHandler) Me(c *gin.Context) {
	u, err := h.svc.GetProfile(c.GetUint("user_id"))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": u})
}

func (h *UserHandler) UpdateMe(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, err := h.svc.UpdateProfile(c.GetUint("user_id"), req.Name)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": u})
}

func (h *UserHandler) List(c *gin.Context) {
	users, err := h.svc.List()
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": users})
}

func (h *UserHandler) UpdateRole(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, err := h.svc.UpdateRole(uint(id), req.Role)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": u})
}

func (h *UserHandler) UpdateStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		IsActive *bool `json:"is_active" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, err := h.svc.SetActive(uint(id), *req.IsActive)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": u})
}