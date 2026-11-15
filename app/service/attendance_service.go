package service

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/ochadella/pbl-peminjaman-ruang/app/model"
	"github.com/ochadella/pbl-peminjaman-ruang/app/repository"
	"github.com/ochadella/pbl-peminjaman-ruang/helper"
	"github.com/ochadella/pbl-peminjaman-ruang/middleware"
)

// WIB dipakai untuk membaca tanggal dan jam operasional perpustakaan.
var WIB = time.FixedZone("WIB", 7*60*60)

type AttendanceService struct {
	bookings   *repository.BookingRepository
	rooms      *repository.RoomRepository
	hours      *repository.OperatingHourRepository
	violations *repository.ViolationRepository
}

func NewAttendanceService(
	bookings *repository.BookingRepository,
	rooms *repository.RoomRepository,
	hours *repository.OperatingHourRepository,
	violations *repository.ViolationRepository,
) *AttendanceService {
	return &AttendanceService{bookings: bookings, rooms: rooms, hours: hours, violations: violations}
}

func isStaff(c *fiber.Ctx) bool {
	role, _ := c.Locals(middleware.LocalRole).(string)
	return role == model.RolePetugas || role == model.RoleAdmin
}

// canHandle: pemilik booking, petugas, atau admin.
func canHandle(c *fiber.Ctx, b model.Booking) bool {
	return b.UserID == middleware.CurrentUserID(c) || isStaff(c)
}

func bookingError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, repository.ErrBookingNotFound):
		return helper.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	case errors.Is(err, repository.ErrBookingStatusChanged):
		return helper.ErrorResponse(c, fiber.StatusConflict, err.Error())
	}
	return helper.HandleError(c, err)
}

// atClock menggabungkan tanggal dengan jam "HH:MM" dalam zona WIB.
func atClock(date time.Time, hhmm string) (time.Time, error) {
	t, err := time.Parse("15:04", hhmm)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), 0, 0, WIB), nil
}

func clock(t time.Time) string {
	return t.In(WIB).Format("15:04")
}

// Availability GET /api/v1/rooms/:id/availability?date=2026-11-20
// Menampilkan jam yang sudah dibooking dan jam yang masih kosong.
func (s *AttendanceService) Availability(c *fiber.Ctx) error {
	id, ok := parseIDParam(c)
	if !ok {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "id ruang tidak valid")
	}
	date, err := time.ParseInLocation("2006-01-02", c.Query("date"), WIB)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "query date wajib diisi dengan format YYYY-MM-DD")
	}

	ctx := c.UserContext()
	room, err := s.rooms.FindByID(ctx, id)
	if err == nil && !room.IsActive && !isAdmin(c) {
		err = repository.ErrRoomNotFound
	}
	if err != nil {
		return roomError(c, err)
	}

	result := model.RoomAvailability{
		RoomID: room.ID, Date: date.Format("2006-01-02"),
		Booked: []model.TimeSlot{}, Free: []model.TimeSlot{},
	}
	hour, err := s.hours.FindByDay(ctx, int(date.Weekday()))
	if errors.Is(err, repository.ErrOperatingHourNotFound) || (err == nil && !hour.IsOpen) {
		return helper.SuccessResponse(c, "perpustakaan tutup pada tanggal ini", result)
	}
	if err != nil {
		return helper.HandleError(c, err)
	}

	open, errOpen := atClock(date, hour.OpenTime)
	closeAt, errClose := atClock(date, hour.CloseTime)
	if errOpen != nil || errClose != nil {
		return helper.HandleError(c, errors.Join(errOpen, errClose))
	}
	result.IsOpen, result.OpenTime, result.CloseTime = true, hour.OpenTime, hour.CloseTime

	bookings, err := s.bookings.ListActiveByRoomBetween(ctx, room.ID, open, closeAt)
	if err != nil {
		return helper.HandleError(c, err)
	}

	cursor := open
	for _, b := range bookings {
		start, end := b.StartAt, b.EndAt
		if start.Before(open) {
			start = open
		}
		if end.After(closeAt) {
			end = closeAt
		}
		result.Booked = append(result.Booked, model.TimeSlot{Start: clock(start), End: clock(end)})
		if start.After(cursor) {
			result.Free = append(result.Free, model.TimeSlot{Start: clock(cursor), End: clock(start)})
		}
		if end.After(cursor) {
			cursor = end
		}
	}
	if closeAt.After(cursor) {
		result.Free = append(result.Free, model.TimeSlot{Start: clock(cursor), End: clock(closeAt)})
	}
	return helper.SuccessResponse(c, "ketersediaan ruang berhasil diambil", result)
}

// CheckIn POST /api/v1/attendance/check-in
// Body berisi kode hasil scan barcode. Hanya bisa 15 menit sebelum s/d 15 menit sesudah jam mulai.
func (s *AttendanceService) CheckIn(c *fiber.Ctx) error {
	var req model.CheckInRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "body request tidak valid")
	}
	code := strings.ToUpper(strings.TrimSpace(req.CheckinCode))
	if code == "" {
		return helper.ErrorResponse(c, fiber.StatusUnprocessableEntity, "checkin_code wajib diisi")
	}

	ctx := c.UserContext()
	b, err := s.bookings.FindByCheckinCode(ctx, code)
	if err == nil && !canHandle(c, b) {
		err = repository.ErrBookingNotFound
	}
	if err != nil {
		return bookingError(c, err)
	}

	switch b.Status {
	case model.BookingBooked:
	case model.BookingCheckedIn:
		return helper.ErrorResponse(c, fiber.StatusConflict, "booking ini sudah check-in")
	default:
		return helper.ErrorResponse(c, fiber.StatusConflict, "booking dengan status "+b.Status+" tidak bisa check-in")
	}

	now := time.Now()
	if !CanCheckIn(b.StartAt, now) {
		if now.Before(b.StartAt) {
			return helper.ErrorResponse(c, fiber.StatusUnprocessableEntity,
				"check-in baru bisa dilakukan mulai jam "+clock(b.StartAt.Add(-CheckInWindow)))
		}
		return helper.ErrorResponse(c, fiber.StatusUnprocessableEntity,
			"waktu check-in sudah lewat (batas jam "+clock(b.StartAt.Add(CheckInWindow))+")")
	}

	b, err = s.bookings.MarkCheckedIn(ctx, b.ID, now)
	if err != nil {
		return bookingError(c, err)
	}
	return helper.SuccessResponse(c, "check-in berhasil", b)
}

// CheckOut POST /api/v1/attendance/:id/check-out
// Lebih dari 15 menit setelah jam selesai otomatis tercatat pelanggaran telat_checkout.
func (s *AttendanceService) CheckOut(c *fiber.Ctx) error {
	id, ok := parseIDParam(c)
	if !ok {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "id booking tidak valid")
	}

	ctx := c.UserContext()
	b, err := s.bookings.FindByID(ctx, id)
	if err == nil && !canHandle(c, b) {
		err = repository.ErrBookingNotFound
	}
	if err != nil {
		return bookingError(c, err)
	}
	if b.Status != model.BookingCheckedIn {
		return helper.ErrorResponse(c, fiber.StatusConflict, "check-out hanya untuk booking yang sudah check-in")
	}

	now := time.Now()
	b, err = s.bookings.MarkCheckedOut(ctx, b.ID, now)
	if err != nil {
		return bookingError(c, err)
	}

	late := IsLateCheckout(b.EndAt, now)
	if late {
		note := "check-out jam " + clock(now) + ", melewati batas jam " + clock(b.EndAt.Add(LateCheckoutTolerance))
		if err := s.recordViolation(ctx, b, model.ViolationLateCheckout, note); err != nil {
			return helper.HandleError(c, err)
		}
	}
	return helper.SuccessResponse(c, "check-out berhasil", fiber.Map{"booking": b, "late": late})
}

// recordViolation mencatat pelanggaran otomatis (recorded_by kosong)
// lalu memblokir mahasiswa bila jumlah pelanggarannya mencapai batas.
func (s *AttendanceService) recordViolation(ctx context.Context, b model.Booking, violationType, note string) error {
	bookingID := b.ID
	_, err := s.violations.Create(ctx, model.CreateViolationRequest{
		UserID: b.UserID, BookingID: &bookingID, Type: violationType, Note: note,
	}, nil)
	if errors.Is(err, repository.ErrViolationExists) {
		return nil
	}
	if err != nil {
		return err
	}

	total, err := s.violations.CountByUser(ctx, b.UserID)
	if err != nil {
		return err
	}
	if until := BlockedUntil(total, time.Now()); until != nil {
		return s.violations.SetBlockedUntil(ctx, b.UserID, *until)
	}
	return nil
}

// ProcessNoShows mengubah booking yang lewat 15 menit tanpa check-in menjadi
// no_show dan mencatat pelanggarannya. Mengembalikan jumlah booking yang diproses.
func (s *AttendanceService) ProcessNoShows(ctx context.Context) (int, error) {
	bookings, err := s.bookings.MarkNoShows(ctx, time.Now().Add(-CheckInWindow))
	if err != nil {
		return 0, err
	}
	for _, b := range bookings {
		note := "tidak check-in sampai jam " + clock(b.StartAt.Add(CheckInWindow))
		if err := s.recordViolation(ctx, b, model.ViolationNoShow, note); err != nil {
			return 0, err
		}
	}
	return len(bookings), nil
}

// RunNoShow POST /api/v1/attendance/process-no-show (petugas/admin).
// Menjalankan pengecekan no-show saat itu juga tanpa menunggu worker.
func (s *AttendanceService) RunNoShow(c *fiber.Ctx) error {
	n, err := s.ProcessNoShows(c.UserContext())
	if err != nil {
		return helper.HandleError(c, err)
	}
	return helper.SuccessResponse(c, "pengecekan no-show selesai", fiber.Map{"processed": n})
}

// StartNoShowWorker menjalankan ProcessNoShows setiap interval di background.
func (s *AttendanceService) StartNoShowWorker(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			if n, err := s.ProcessNoShows(ctx); err != nil {
				log.Printf("no-show worker: %v", err)
			} else if n > 0 {
				log.Printf("no-show worker: %d booking diubah menjadi no_show", n)
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}
