package service

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/ochadella/pbl-peminjaman-ruang/app/model"
	"github.com/ochadella/pbl-peminjaman-ruang/app/repository"
	"github.com/ochadella/pbl-peminjaman-ruang/helper"
	"github.com/ochadella/pbl-peminjaman-ruang/middleware"
)

type ViolationService struct {
	violations *repository.ViolationRepository
	users      *repository.UserRepository
}

func NewViolationService(violations *repository.ViolationRepository, users *repository.UserRepository) *ViolationService {
	return &ViolationService{violations: violations, users: users}
}

func validViolationType(t string) bool {
	switch t {
	case model.ViolationNoShow, model.ViolationLateCheckout, model.ViolationDamage:
		return true
	}
	return false
}

// Create POST /api/v1/violations (petugas/admin).
// Setelah dicatat, bila total pelanggaran mencapai batas, mahasiswa diblokir booking.
func (s *ViolationService) Create(c *fiber.Ctx) error {
	var req model.CreateViolationRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "body request tidak valid")
	}
	req.Type = strings.TrimSpace(req.Type)
	req.Note = strings.TrimSpace(req.Note)
	switch {
	case req.UserID <= 0:
		return helper.ErrorResponse(c, fiber.StatusUnprocessableEntity, "user_id wajib diisi")
	case !validViolationType(req.Type):
		return helper.ErrorResponse(c, fiber.StatusUnprocessableEntity, "type harus no_show, telat_checkout, atau kerusakan")
	case len(req.Note) > 500:
		return helper.ErrorResponse(c, fiber.StatusUnprocessableEntity, "catatan maksimal 500 karakter")
	}

	ctx := c.UserContext()
	target, err := s.users.FindByID(ctx, req.UserID)
	if errors.Is(err, repository.ErrUserNotFound) {
		return helper.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	if err != nil {
		return helper.HandleError(c, err)
	}
	if target.Role != model.RoleMahasiswa {
		return helper.ErrorResponse(c, fiber.StatusUnprocessableEntity, "pelanggaran hanya bisa dicatat untuk mahasiswa")
	}

	recordedBy := middleware.CurrentUserID(c)
	violation, err := s.violations.Create(ctx, req, &recordedBy)
	var pgErr *pgconn.PgError
	switch {
	case errors.Is(err, repository.ErrViolationExists):
		return helper.ErrorResponse(c, fiber.StatusConflict, err.Error())
	case errors.As(err, &pgErr) && pgErr.Code == "23503":
		return helper.ErrorResponse(c, fiber.StatusNotFound, "booking tidak ditemukan")
	case err != nil:
		return helper.HandleError(c, err)
	}

	total, err := s.violations.CountByUser(ctx, req.UserID)
	if err != nil {
		return helper.HandleError(c, err)
	}
	blockedUntil := BlockedUntil(total, time.Now())
	if blockedUntil != nil {
		if err := s.violations.SetBlockedUntil(ctx, req.UserID, *blockedUntil); err != nil {
			return helper.HandleError(c, err)
		}
	}

	return c.Status(fiber.StatusCreated).JSON(helper.Response{
		Success: true,
		Message: "pelanggaran berhasil dicatat",
		Data: fiber.Map{
			"violation":        violation,
			"total_violations": total,
			"blocked_until":    blockedUntil,
		},
	})
}

// ListMine GET /api/v1/violations/me. Riwayat pelanggaran user yang login.
func (s *ViolationService) ListMine(c *fiber.Ctx) error {
	list, err := s.violations.ListByUser(c.UserContext(), middleware.CurrentUserID(c))
	if err != nil {
		return helper.HandleError(c, err)
	}
	return helper.SuccessResponse(c, "riwayat pelanggaran berhasil diambil", list)
}

// ListByUser GET /api/v1/violations/user/:id (petugas/admin).
func (s *ViolationService) ListByUser(c *fiber.Ctx) error {
	id, ok := parseIDParam(c)
	if !ok {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "id user tidak valid")
	}
	list, err := s.violations.ListByUser(c.UserContext(), id)
	if err != nil {
		return helper.HandleError(c, err)
	}
	return helper.SuccessResponse(c, "riwayat pelanggaran berhasil diambil", list)
}
