package service

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ochadella/pbl-peminjaman-ruang/app/model"
	"github.com/ochadella/pbl-peminjaman-ruang/app/repository"
	"github.com/ochadella/pbl-peminjaman-ruang/helper"
	"github.com/ochadella/pbl-peminjaman-ruang/middleware"

	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type BookingService struct {
	db          *pgxpool.Pool
	bookingRepo *repository.BookingRepository
	userRepo    *repository.UserRepository
	roomRepo    *repository.RoomRepository
	hourRepo    *repository.OperatingHourRepository
}

func NewBookingService(
	db *pgxpool.Pool,
	bookingRepo *repository.BookingRepository,
	userRepo *repository.UserRepository,
	roomRepo *repository.RoomRepository,
	hourRepo *repository.OperatingHourRepository,
) *BookingService {
	return &BookingService{
		db:          db,
		bookingRepo: bookingRepo,
		userRepo:    userRepo,
		roomRepo:    roomRepo,
		hourRepo:    hourRepo,
	}
}

// CreateBooking membuat booking baru untuk mahasiswa yang sedang login.
func (s *BookingService) CreateBooking(c *fiber.Ctx) error {
	userID := middleware.CurrentUserID(c)

	if userID == 0 {
		return helper.ErrorResponse(
			c,
			fiber.StatusUnauthorized,
			"silakan login terlebih dahulu",
		)
	}

	var req model.CreateBookingRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(
			c,
			fiber.StatusBadRequest,
			"format request tidak valid",
		)
	}

	// Periksa data mahasiswa.
	user, err := s.userRepo.FindByID(c.UserContext(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "akun tidak ditemukan")
		}
		return helper.HandleError(c, err)
	}

	if !user.IsActive {
		return helper.ErrorResponse(c, fiber.StatusForbidden, "akun mahasiswa tidak aktif")
	}

	if user.Role != model.RoleMahasiswa {
		return helper.ErrorResponse(c, fiber.StatusForbidden, "hanya mahasiswa yang dapat membuat booking")
	}

	if user.BlockedUntil != nil && user.BlockedUntil.After(time.Now()) {
		return helper.ErrorResponse(c, fiber.StatusForbidden, "akun sedang diblokir dari booking")
	}

	// Periksa data ruangan.
	room, err := s.roomRepo.FindByID(c.UserContext(), req.RoomID)
	if err != nil {
		if errors.Is(err, repository.ErrRoomNotFound) {
			return helper.ErrorResponse(c, fiber.StatusNotFound, "ruangan tidak ditemukan")
		}
		return helper.HandleError(c, err)
	}

	if !room.IsActive {
		return helper.ErrorResponse(c, fiber.StatusForbidden, "ruangan sedang tidak aktif")
	}

	if req.ParticipantCount <= 0 || req.ParticipantCount > room.Capacity {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "jumlah peserta melebihi kapasitas ruangan atau tidak valid")
	}

	// Validasi tujuan booking.
	req.Purpose = strings.TrimSpace(req.Purpose)
	if req.Purpose == "" || len([]rune(req.Purpose)) > 200 {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "tujuan booking wajib diisi maksimal 200 karakter")
	}

	// Gunakan zona waktu WIB.
	now := time.Now().In(WIB)
	startAt := req.StartAt.In(WIB)
	endAt := req.EndAt.In(WIB)

	if !startAt.After(now) {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "waktu mulai booking harus setelah waktu sekarang")
	}

	if !endAt.After(startAt) {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "waktu selesai harus setelah waktu mulai")
	}

	if startAt.Format("2006-01-02") != endAt.Format("2006-01-02") {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "booking harus dimulai dan selesai pada tanggal yang sama")
	}

	if endAt.Sub(startAt) > 2*time.Hour {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "durasi booking maksimal 2 jam")
	}

	// Periksa jam operasional berdasarkan hari booking.
	operatingHour, err := s.hourRepo.FindByDay(
		c.UserContext(),
		int(startAt.Weekday()),
	)
	if err != nil {
		if errors.Is(err, repository.ErrOperatingHourNotFound) {
			return helper.ErrorResponse(c, fiber.StatusBadRequest, "jam operasional belum tersedia")
		}
		return helper.HandleError(c, err)
	}

	if !operatingHour.IsOpen {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "perpustakaan tutup pada hari tersebut")
	}

	openAt, err := atClock(startAt, operatingHour.OpenTime)
	if err != nil {
		return helper.HandleError(c, err)
	}

	closeAt, err := atClock(startAt, operatingHour.CloseTime)
	if err != nil {
		return helper.HandleError(c, err)
	}

	if startAt.Before(openAt) || endAt.After(closeAt) {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "booking harus berada dalam jam operasional perpustakaan")
	}

	// Mulai transaksi untuk memastikan validasi dan penyimpanan konsisten.
	ctx := c.UserContext()

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return helper.HandleError(c, err)
	}
	defer tx.Rollback(ctx)

	txBookingRepo := repository.NewBookingRepository(tx)

	// Kunci berdasarkan mahasiswa dan tanggal booking.
	dayStart := time.Date(
		startAt.Year(), startAt.Month(), startAt.Day(),
		0, 0, 0, 0, WIB,
	)
	dayEnd := dayStart.AddDate(0, 0, 1)

	userLock := fmt.Sprintf(
		"booking-user:%d:%s",
		userID,
		dayStart.Format("2006-01-02"),
	)
	roomLock := fmt.Sprintf(
		"booking-room:%d:%s",
		req.RoomID,
		dayStart.Format("2006-01-02"),
	)

	if err := acquireBookingLock(ctx, tx, userLock); err != nil {
		return helper.HandleError(c, err)
	}

	if err := acquireBookingLock(ctx, tx, roomLock); err != nil {
		return helper.HandleError(c, err)
	}

	// Periksa apakah mahasiswa sudah menggunakan jatah booking hari ini.
	hasDailyBooking, err := txBookingRepo.HasDailyBooking(
		ctx,
		userID,
		dayStart,
		dayEnd,
	)
	if err != nil {
		return helper.HandleError(c, err)
	}

	if hasDailyBooking {
		return helper.ErrorResponse(
			c,
			fiber.StatusConflict,
			"mahasiswa sudah menggunakan jatah booking hari ini",
		)
	}

	// Periksa apakah jadwal ruangan bertabrakan.
	hasConflict, err := txBookingRepo.HasRoomConflict(
		ctx,
		req.RoomID,
		startAt,
		endAt,
	)
	if err != nil {
		return helper.HandleError(c, err)
	}

	if hasConflict {
		return helper.ErrorResponse(
			c,
			fiber.StatusConflict,
			"jadwal ruangan bertabrakan dengan booking lain",
		)
	}

	// Buat kode check-in untuk booking.
	checkinCode, err := generateCheckinCode()
	if err != nil {
		return helper.HandleError(c, err)
	}

	// Simpan booking ke database.
	booking, err := txBookingRepo.Create(
		ctx,
		userID,
		req.RoomID,
		startAt,
		endAt,
		req.Purpose,
		req.ParticipantCount,
		checkinCode,
	)
	if err != nil {
		if errors.Is(err, repository.ErrBookingConflict) {
			return helper.ErrorResponse(
				c,
				fiber.StatusConflict,
				"jadwal booking bertabrakan",
			)
		}
		return helper.HandleError(c, err)
	}

	// Commit transaksi.
	if err := tx.Commit(ctx); err != nil {
		return helper.HandleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(helper.Response{
		Success: true,
		Message: "booking berhasil dibuat",
		Data:    booking,
	})
}

// acquireBookingLock mengunci proses booking agar validasi tidak berjalan
// bersamaan untuk mahasiswa atau ruangan yang sama.
func acquireBookingLock(
	ctx context.Context,
	tx pgx.Tx,
	key string,
) error {
	_, err := tx.Exec(
		ctx,
		`SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`,
		key,
	)
	return err
}

// generateCheckinCode membuat kode check-in acak sepanjang 12 karakter.
func generateCheckinCode() (string, error) {
	bytes := make([]byte, 6)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return strings.ToUpper(hex.EncodeToString(bytes)), nil
}

// ListMine menampilkan seluruh booking milik mahasiswa yang sedang login.
func (s *BookingService) ListMine(c *fiber.Ctx) error {
	userID := middleware.CurrentUserID(c)

	if userID == 0 {
		return helper.ErrorResponse(
			c,
			fiber.StatusUnauthorized,
			"silakan login terlebih dahulu",
		)
	}

	bookings, err := s.bookingRepo.ListByUser(c.UserContext(), userID)
	if err != nil {
		return helper.HandleError(c, err)
	}

	return helper.SuccessResponse(
		c,
		"daftar booking berhasil diambil",
		bookings,
	)
}

// CancelMine membatalkan booking milik mahasiswa yang sedang login.
func (s *BookingService) CancelMine(c *fiber.Ctx) error {
	userID := middleware.CurrentUserID(c)

	if userID == 0 {
		return helper.ErrorResponse(
			c,
			fiber.StatusUnauthorized,
			"silakan login terlebih dahulu",
		)
	}

	bookingID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || bookingID <= 0 {
		return helper.ErrorResponse(
			c,
			fiber.StatusBadRequest,
			"ID booking tidak valid",
		)
	}

	booking, err := s.bookingRepo.FindByID(c.UserContext(), bookingID)
	if err != nil {
		if errors.Is(err, repository.ErrBookingNotFound) {
			return helper.ErrorResponse(
				c,
				fiber.StatusNotFound,
				"booking tidak ditemukan",
			)
		}
		return helper.HandleError(c, err)
	}

	if booking.UserID != userID {
		return helper.ErrorResponse(
			c,
			fiber.StatusNotFound,
			"booking tidak ditemukan",
		)
	}

	if booking.Status != model.BookingBooked {
		return helper.ErrorResponse(
			c,
			fiber.StatusConflict,
			"booking tidak dapat dibatalkan",
		)
	}

	now := time.Now()
	if now.After(booking.StartAt.Add(-30 * time.Minute)) {
		return helper.ErrorResponse(
			c,
			fiber.StatusBadRequest,
			"booking hanya dapat dibatalkan minimal 30 menit sebelum waktu mulai",
		)
	}

	cancelledBooking, err := s.bookingRepo.Cancel(
		c.UserContext(),
		bookingID,
		userID,
		now,
	)

	if err != nil {
		if errors.Is(err, repository.ErrBookingStatusChanged) {
			return helper.ErrorResponse(
				c,
				fiber.StatusConflict,
				"status booking sudah berubah, silakan muat ulang",
			)
		}
		return helper.HandleError(c, err)
	}

	return helper.SuccessResponse(
		c,
		"booking berhasil dibatalkan",
		cancelledBooking,
	)
}
