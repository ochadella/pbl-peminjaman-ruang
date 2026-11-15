package service

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/ochadella/pbl-peminjaman-ruang/app/model"
	"github.com/ochadella/pbl-peminjaman-ruang/app/repository"
	"github.com/ochadella/pbl-peminjaman-ruang/helper"
)

type OperatingHourService struct {
	hours *repository.OperatingHourRepository
}

func NewOperatingHourService(hours *repository.OperatingHourRepository) *OperatingHourService {
	return &OperatingHourService{hours: hours}
}

// List GET /api/v1/operating-hours.
func (s *OperatingHourService) List(c *fiber.Ctx) error {
	hours, err := s.hours.List(c.UserContext())
	if err != nil {
		return helper.HandleError(c, err)
	}
	return helper.SuccessResponse(c, "jam operasional berhasil diambil", hours)
}

// Update PUT /api/v1/operating-hours/:day (admin). day: 0 = Minggu ... 6 = Sabtu.
func (s *OperatingHourService) Update(c *fiber.Ctx) error {
	day, err := strconv.Atoi(c.Params("day"))
	if err != nil || day < 0 || day > 6 {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "hari harus angka 0 (Minggu) sampai 6 (Sabtu)")
	}
	var req model.UpdateOperatingHourRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "body request tidak valid")
	}
	open, errOpen := time.Parse("15:04", req.OpenTime)
	closeAt, errClose := time.Parse("15:04", req.CloseTime)
	if errOpen != nil || errClose != nil {
		return helper.ErrorResponse(c, fiber.StatusUnprocessableEntity, "format jam harus HH:MM, contoh 08:00")
	}
	if req.IsOpen && !closeAt.After(open) {
		return helper.ErrorResponse(c, fiber.StatusUnprocessableEntity, repository.ErrInvalidOperatingHour.Error())
	}

	hour, err := s.hours.Upsert(c.UserContext(), day, req)
	if errors.Is(err, repository.ErrInvalidOperatingHour) {
		return helper.ErrorResponse(c, fiber.StatusUnprocessableEntity, err.Error())
	}
	if err != nil {
		return helper.HandleError(c, err)
	}
	return helper.SuccessResponse(c, "jam operasional berhasil diubah", hour)
}
