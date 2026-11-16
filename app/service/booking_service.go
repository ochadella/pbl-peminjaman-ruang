package service

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/ochadella/pbl-peminjaman-ruang/app/model"
	"github.com/ochadella/pbl-peminjaman-ruang/app/repository"
	"github.com/ochadella/pbl-peminjaman-ruang/helper"
	"github.com/ochadella/pbl-peminjaman-ruang/middleware"
)

type BookingService struct {
	bookings *repository.BookingRepository
	rooms    *repository.RoomRepository
	hours    *repository.OperatingHourRepository
	users    *repository.UserRepository
}

func NewBookingService(
	bookings *repository.BookingRepository,
	rooms *repository.RoomRepository,
	hours *repository.OperatingHourRepository,
	users *repository.UserRepository,
) *BookingService {
	return &BookingService{bookings: bookings, rooms: rooms, hours: hours, users: users}
}

func bookingWriteError(c *fiber.Ctx, err error) error {
	if errors.Is(err, repository.ErrBookingConflict) {
		return helper.ErrorResponse(c, fiber.StatusConflict, err.Error())
	}
	return bookingError(c, err)
}

func (s *BookingService) Create(c *fiber.Ctx) error {
	var req model.CreateBookingRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest,
			"body request tidak valid (start_at dan end_at memakai format RFC3339, contoh 2026-11-20T10:00:00+07:00)")
	}
	req.Purpose = strings.TrimSpace(req.Purpose)
	if req.RoomID <= 0 || req.Purpose == "" || len(req.Purpose) > 200 || req.ParticipantCount <= 0 {
		return helper.ErrorResponse(c, fiber.StatusUnprocessableEntity,
			"room_id, purpose (maksimal 200 karakter), dan participant_count wajib diisi dengan benar")
	}
	now := time.Now()
	if err := ValidateBookingTime(req.StartAt, req.EndAt, now); err != nil {
		return helper.ErrorResponse(c, fiber.StatusUnprocessableEntity, err.Error())
	}

	ctx := c.UserContext()
	userID := middleware.CurrentUserID(c)

	user, err := s.users.FindByID(ctx, userID)
	if errors.Is(err, repository.ErrUserNotFound) {
		return helper.ErrorResponse(c, fiber.StatusUnauthorized, "akun tidak ditemukan")
	}
	if err != nil {
		return helper.HandleError(c, err)
	}
	if !user.IsActive {
		return helper.ErrorResponse(c, fiber.StatusForbidden, "akun tidak aktif")
	}
	if user.BlockedUntil != nil && user.BlockedUntil.After(now) {
		return helper.ErrorResponse(c, fiber.StatusForbidden,
			"anda diblokir booking sampai "+user.BlockedUntil.In(WIB).Format("02-01-2006 15:04")+" WIB")
	}

	room, err := s.rooms.FindByID(ctx, req.RoomID)
	if err == nil && !room.IsActive {
		err = repository.ErrRoomNotFound
	}
	if err != nil {
		return roomError(c, err)
	}
	if req.ParticipantCount > room.Capacity {
		return helper.ErrorResponse(c, fiber.StatusUnprocessableEntity,
			"jumlah peserta melebihi kapasitas ruang (maksimal "+strconv.Itoa(room.Capacity)+")")
	}

	startWIB := req.StartAt.In(WIB)
	dayStart := time.Date(startWIB.Year(), startWIB.Month(), startWIB.Day(), 0, 0, 0, 0, WIB)
	hour, err := s.hours.FindByDay(ctx, int(startWIB.Weekday()))
	if errors.Is(err, repository.ErrOperatingHourNotFound) || (err == nil && !hour.IsOpen) {
		return helper.ErrorResponse(c, fiber.StatusUnprocessableEntity, "perpustakaan tutup pada tanggal tersebut")
	}
	if err != nil {
		return helper.HandleError(c, err)
	}
	open, errOpen := atClock(dayStart, hour.OpenTime)
	closeAt, errClose := atClock(dayStart, hour.CloseTime)
	if errOpen != nil || errClose != nil {
		return helper.HandleError(c, errors.Join(errOpen, errClose))
	}
	if !WithinOperatingHours(req.StartAt, req.EndAt, open, closeAt) {
		return helper.ErrorResponse(c, fiber.StatusUnprocessableEntity,
			"booking harus di dalam jam operasional ("+hour.OpenTime+" sampai "+hour.CloseTime+")")
	}

	count, err := s.bookings.CountActiveByUserBetween(ctx, userID, dayStart, dayStart.AddDate(0, 0, 1))
	if err != nil {
		return helper.HandleError(c, err)
	}
	if count >= 1 {
		return helper.ErrorResponse(c, fiber.StatusConflict, "maksimal 1 booking aktif per hari")
	}

	booking := model.Booking{
		UserID: userID, RoomID: req.RoomID, StartAt: req.StartAt, EndAt: req.EndAt,
		Purpose: req.Purpose, ParticipantCount: req.ParticipantCount,
	}
	for attempt := 0; attempt < 5; attempt++ {
		code, err := NewCheckinCode()
		if err != nil {
			return helper.HandleError(c, err)
		}
		created, err := s.bookings.Create(ctx, booking, code)
		if errors.Is(err, repository.ErrCheckinCodeTaken) {
			continue 
		}
		if err != nil {
			return bookingWriteError(c, err)
		}
		return c.Status(fiber.StatusCreated).JSON(helper.Response{
			Success: true, Message: "booking berhasil dibuat", Data: created,
		})
	}
	return helper.ErrorResponse(c, fiber.StatusInternalServerError, "gagal membuat kode check-in, coba lagi")
}

func (s *BookingService) ListMine(c *fiber.Ctx) error {
	list, err := s.bookings.ListByUser(c.UserContext(), middleware.CurrentUserID(c))
	if err != nil {
		return helper.HandleError(c, err)
	}
	return helper.SuccessResponse(c, "riwayat booking berhasil diambil", list)
}

func (s *BookingService) Get(c *fiber.Ctx) error {
	id, ok := parseIDParam(c)
	if !ok {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "id booking tidak valid")
	}
	b, err := s.bookings.FindByID(c.UserContext(), id)
	if err == nil && !canHandle(c, b) {
		err = repository.ErrBookingNotFound
	}
	if err != nil {
		return bookingError(c, err)
	}
	return helper.SuccessResponse(c, "booking berhasil diambil", b)
}

func (s *BookingService) Cancel(c *fiber.Ctx) error {
	id, ok := parseIDParam(c)
	if !ok {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "id booking tidak valid")
	}
	ctx := c.UserContext()
	b, err := s.bookings.FindByID(ctx, id)
	if err == nil && b.UserID != middleware.CurrentUserID(c) {
		err = repository.ErrBookingNotFound
	}
	if err != nil {
		return bookingError(c, err)
	}
	if b.Status != model.BookingBooked {
		return helper.ErrorResponse(c, fiber.StatusConflict, "hanya booking berstatus booked yang bisa dibatalkan")
	}
	now := time.Now()
	if !CanCancel(b.StartAt, now) {
		return helper.ErrorResponse(c, fiber.StatusUnprocessableEntity, "pembatalan minimal 30 menit sebelum jam mulai")
	}
	b, err = s.bookings.Cancel(ctx, b.ID, now)
	if err != nil {
		return bookingError(c, err)
	}
	return helper.SuccessResponse(c, "booking berhasil dibatalkan", b)
}