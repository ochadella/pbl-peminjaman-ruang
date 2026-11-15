package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ochadella/pbl-peminjaman-ruang/helper"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin", AuthMiddleware(), RoleMiddleware("admin"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return r
}

func doRequest(r *gin.Engine, token string) int {
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code
}

func TestMiddleware(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r := setupRouter()

	assert.Equal(t, http.StatusUnauthorized, doRequest(r, ""))

	mhs, _ := helper.GenerateToken(1, "mahasiswa")
	assert.Equal(t, http.StatusForbidden, doRequest(r, mhs))

	adm, _ := helper.GenerateToken(2, "admin")
	assert.Equal(t, http.StatusOK, doRequest(r, adm))
}