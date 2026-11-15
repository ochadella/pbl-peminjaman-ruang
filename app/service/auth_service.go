package service

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"

	"github.com/ochadella/pbl-peminjaman-ruang/app/model"
	"github.com/ochadella/pbl-peminjaman-ruang/app/repository"
	"github.com/ochadella/pbl-peminjaman-ruang/helper"
	"github.com/ochadella/pbl-peminjaman-ruang/middleware"
)

type AuthService struct {
	users     *repository.UserRepository
	jwtSecret string
}

func NewAuthService(users *repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{users: users, jwtSecret: jwtSecret}
}

func validateRegister(req model.RegisterRequest) error {
	switch {
	case strings.TrimSpace(req.NimNip) == "" || len(req.NimNip) > 20:
		return errors.New("NIM/NIP wajib diisi, maksimal 20 karakter")
	case strings.TrimSpace(req.Name) == "":
		return errors.New("nama wajib diisi")
	case !strings.Contains(req.Email, "@"):
		return errors.New("format email tidak valid")
	case len(req.Password) < 8:
		return errors.New("password minimal 8 karakter")
	}
	return nil
}

// Register POST /api/v1/auth/register. Akun baru selalu ber-role mahasiswa.
func (s *AuthService) Register(c *fiber.Ctx) error {
	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "body request tidak valid")
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if err := validateRegister(req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusUnprocessableEntity, err.Error())
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return helper.HandleError(c, err)
	}
	user, err := s.users.Create(c.UserContext(), req.NimNip, req.Name, req.Email, string(hash))
	if errors.Is(err, repository.ErrUserDuplicate) {
		return helper.ErrorResponse(c, fiber.StatusConflict, err.Error())
	}
	if err != nil {
		return helper.HandleError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(helper.Response{
		Success: true, Message: "registrasi berhasil", Data: user,
	})
}

// Login POST /api/v1/auth/login. Mengembalikan token JWT.
func (s *AuthService) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "body request tidak valid")
	}
	user, err := s.users.FindByEmail(c.UserContext(), strings.ToLower(strings.TrimSpace(req.Email)))
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return helper.ErrorResponse(c, fiber.StatusUnauthorized, "email atau password salah")
	}
	if !user.IsActive {
		return helper.ErrorResponse(c, fiber.StatusForbidden, "akun tidak aktif")
	}
	token, err := helper.GenerateToken(s.jwtSecret, user.ID, user.Role)
	if err != nil {
		return helper.HandleError(c, err)
	}
	return helper.SuccessResponse(c, "login berhasil", model.LoginResponse{Token: token, User: user})
}

// Me GET /api/v1/auth/me. Profil user yang sedang login.
func (s *AuthService) Me(c *fiber.Ctx) error {
	user, err := s.users.FindByID(c.UserContext(), middleware.CurrentUserID(c))
	if errors.Is(err, repository.ErrUserNotFound) {
		return helper.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	if err != nil {
		return helper.HandleError(c, err)
	}
	return helper.SuccessResponse(c, "profil berhasil diambil", user)
}
