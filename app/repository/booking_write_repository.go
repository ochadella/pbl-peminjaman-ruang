package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/ochadella/pbl-peminjaman-ruang/app/model"
)

var (
	ErrBookingConflict  = errors.New("ruang sudah dibooking pada jam tersebut")
	ErrCheckinCodeTaken = errors.New("kode check-in sudah dipakai")
)

func (r *BookingRepository) Create(ctx context.Context, b model.Booking, code string) (model.Booking, error) {
	created, err := scanBooking(r.db.QueryRow(ctx,
		`INSERT INTO bookings (user_id, room_id, start_at, end_at, purpose, participant_count, checkin_code)
		 SELECT $1::bigint, $2::bigint, $3::timestamptz, $4::timestamptz, $5::varchar, $6::int, $7::varchar
		 WHERE NOT EXISTS (
		   SELECT 1 FROM bookings
		   WHERE room_id = $2::bigint
		     AND status IN ('booked', 'checked_in', 'completed')
		     AND start_at < $4::timestamptz AND end_at > $3::timestamptz)
		 RETURNING `+bookingColumns,
		b.UserID, b.RoomID, b.StartAt, b.EndAt, b.Purpose, b.ParticipantCount, code))
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return model.Booking{}, ErrBookingConflict
	case isUniqueViolation(err):
		return model.Booking{}, ErrCheckinCodeTaken
	}
	return created, err
}

func (r *BookingRepository) CountActiveByUserBetween(ctx context.Context, userID int64, from, to time.Time) (int, error) {
	var n int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM bookings
		 WHERE user_id = $1 AND status IN ('booked', 'checked_in')
		   AND start_at >= $2 AND start_at < $3`, userID, from, to).Scan(&n)
	return n, err
}

func (r *BookingRepository) ListByUser(ctx context.Context, userID int64) ([]model.Booking, error) {
	return scanBookings(r.db.Query(ctx,
		`SELECT `+bookingColumns+` FROM bookings WHERE user_id = $1 ORDER BY start_at DESC`, userID))
}

func (r *BookingRepository) Cancel(ctx context.Context, id int64, at time.Time) (model.Booking, error) {
	b, err := scanBooking(r.db.QueryRow(ctx,
		`UPDATE bookings SET status = 'cancelled', cancelled_at = $2, updated_at = NOW()
		 WHERE id = $1 AND status = 'booked'
		 RETURNING `+bookingColumns, id, at))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Booking{}, ErrBookingStatusChanged
	}
	return b, err
}