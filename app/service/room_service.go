package service

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/ochadella/pbl-peminjaman-ruang/app/model"
	"github.com/ochadella/pbl-peminjaman-ruang/app/repository"
	"github.com/ochadella/pbl-peminjaman-ruang/helper"
	"github.com/ochadella/pbl-peminjaman-ruang/middleware"
)

type RoomService struct {
	rooms *repository.RoomRepository
}

func NewRoomService(rooms *repository.RoomRepository) *RoomService {
	return &RoomService{rooms: rooms}
}

func parseIDParam(c *fiber.Ctx) (int64, bool) {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	return id, err == nil && id > 0
}

func isAdmin(c *fiber.Ctx) bool {
	role, _ := c.Locals(middleware.LocalRole).(string)
	return role == model.RoleAdmin
}

func roomError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, repository.ErrRoomNotFound):
		return helper.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	case errors.Is(err, repository.ErrRoomCodeTaken):
		return helper.ErrorResponse(c, fiber.StatusConflict, err.Error())
	}
	return helper.HandleError(c, err)
}

// List GET /api/v1/rooms. Admin melihat semua ruang, selain admin hanya ruang aktif.
func (s *RoomService) List(c *fiber.Ctx) error {
	rooms, err := s.rooms.List(c.UserContext(), !isAdmin(c))
	if err != nil {
		return helper.HandleError(c, err)
	}
	return helper.SuccessResponse(c, "daftar ruang berhasil diambil", rooms)
}

// Get GET /api/v1/rooms/:id.
func (s *RoomService) Get(c *fiber.Ctx) error {
	id, ok := parseIDParam(c)
	if !ok {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "id ruang tidak valid")
	}
	room, err := s.rooms.FindByID(c.UserContext(), id)
	if err == nil && !room.IsActive && !isAdmin(c) {
		err = repository.ErrRoomNotFound
	}
	if err != nil {
		return roomError(c, err)
	}
	return helper.SuccessResponse(c, "detail ruang berhasil diambil", room)
}

// Create POST /api/v1/rooms (admin).
func (s *RoomService) Create(c *fiber.Ctx) error {
	var req model.CreateRoomRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "body request tidak valid")
	}
	req.Code = strings.ToUpper(strings.TrimSpace(req.Code))
	req.Name = strings.TrimSpace(req.Name)
	req.Location = strings.TrimSpace(req.Location)

	switch {
	case req.Code == "" || len(req.Code) > 20:
		return helper.ErrorResponse(c, fiber.StatusUnprocessableEntity, "kode ruang wajib diisi, maksimal 20 karakter")
	case req.Name == "" || len(req.Name) > 100:
		return helper.ErrorResponse(c, fiber.StatusUnprocessableEntity, "nama ruang wajib diisi, maksimal 100 karakter")
	case req.Location == "" || len(req.Location) > 100:
		return helper.ErrorResponse(c, fiber.StatusUnprocessableEntity, "lokasi wajib diisi, maksimal 100 karakter")
	case req.Capacity <= 0:
		return helper.ErrorResponse(c, fiber.StatusUnprocessableEntity, "kapasitas harus lebih dari 0")
	}

	room, err := s.rooms.Create(c.UserContext(), req)
	if err != nil {
		return roomError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(helper.Response{
		Success: true, Message: "ruang berhasil ditambahkan", Data: room,
	})
}

// Update PATCH /api/v1/rooms/:id (admin). Field yang tidak dikirim tidak diubah.
func (s *RoomService) Update(c *fiber.Ctx) error {
	id, ok := parseIDParam(c)
	if !ok {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "id ruang tidak valid")
	}
	var req model.UpdateRoomRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "body request tidak valid")
	}
	switch {
	case req.Name != nil && (strings.TrimSpace(*req.Name) == "" || len(*req.Name) > 100):
		return helper.ErrorResponse(c, fiber.StatusUnprocessableEntity, "nama ruang tidak boleh kosong, maksimal 100 karakter")
	case req.Location != nil && (strings.TrimSpace(*req.Location) == "" || len(*req.Location) > 100):
		return helper.ErrorResponse(c, fiber.StatusUnprocessableEntity, "lokasi tidak boleh kosong, maksimal 100 karakter")
	case req.Capacity != nil && *req.Capacity <= 0:
		return helper.ErrorResponse(c, fiber.StatusUnprocessableEntity, "kapasitas harus lebih dari 0")
	}

	room, err := s.rooms.Update(c.UserContext(), id, req)
	if err != nil {
		return roomError(c, err)
	}
	return helper.SuccessResponse(c, "ruang berhasil diubah", room)
}
