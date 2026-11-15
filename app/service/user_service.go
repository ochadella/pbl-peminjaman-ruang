package service

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/ochadella/pbl-peminjaman-ruang/app/model"
	"github.com/ochadella/pbl-peminjaman-ruang/app/repository"
	"github.com/ochadella/pbl-peminjaman-ruang/helper"
)

type UserService struct {
	users *repository.UserRepository
}

func NewUserService(users *repository.UserRepository) *UserService {
	return &UserService{users: users}
}

// List GET /api/v1/users (admin).
func (s *UserService) List(c *fiber.Ctx) error {
	users, err := s.users.List(c.UserContext())
	if err != nil {
		return helper.HandleError(c, err)
	}
	return helper.SuccessResponse(c, "daftar user berhasil diambil", users)
}

// UpdateRole PATCH /api/v1/users/:id/role (admin).
func (s *UserService) UpdateRole(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "id user tidak valid")
	}
	var req model.UpdateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "body request tidak valid")
	}
	switch req.Role {
	case model.RoleMahasiswa, model.RolePetugas, model.RoleAdmin:
	default:
		return helper.ErrorResponse(c, fiber.StatusUnprocessableEntity, "role harus mahasiswa, petugas, atau admin")
	}

	user, err := s.users.UpdateRole(c.UserContext(), id, req.Role)
	if errors.Is(err, repository.ErrUserNotFound) {
		return helper.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	if err != nil {
		return helper.HandleError(c, err)
	}
	return helper.SuccessResponse(c, "role user berhasil diubah", user)
}
