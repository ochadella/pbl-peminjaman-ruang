package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/ochadella/pbl-peminjaman-ruang/app/model"
)

var (
	ErrBookingNotFound      = errors.New("booking tidak ditemukan")
	ErrBookingStatusChanged = errors.New("status booking sudah berubah, coba muat ulang")
)

const bookingColumns = `id, user_id, room_id, start_at, end_at, purpose, participant_count, status,
	checkin_code, checked_in_at, checked_out_at, cancelled_at, created_at, updated_at`

type BookingRepository struct {
	db DBTX
}

func NewBookingRepository(db DBTX) *BookingRepository {
	return &BookingRepository{db: db}
}

func scanBooking(row pgx.Row) (model.Booking, error) {
	var b model.Booking
	err := row.Scan(&b.ID, &b.UserID, &b.RoomID, &b.StartAt, &b.EndAt, &b.Purpose,
		&b.ParticipantCount, &b.Status, &b.CheckinCode, &b.CheckedInAt,
		&b.CheckedOutAt, &b.CancelledAt, &b.CreatedAt, &b.UpdatedAt)
	return b, err
}

func scanBookings(rows pgx.Rows, err error) ([]model.Booking, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []model.Booking{}
	for rows.Next() {
		b, err := scanBooking(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, b)
	}
	return list, rows.Err()
}

func (r *BookingRepository) FindByID(ctx context.Context, id int64) (model.Booking, error) {
	b, err := scanBooking(r.db.QueryRow(ctx, `SELECT `+bookingColumns+` FROM bookings WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Booking{}, ErrBookingNotFound
	}
	return b, err
}

func (r *BookingRepository) FindByCheckinCode(ctx context.Context, code string) (model.Booking, error) {
	b, err := scanBooking(r.db.QueryRow(ctx, `SELECT `+bookingColumns+` FROM bookings WHERE checkin_code = $1`, code))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Booking{}, ErrBookingNotFound
	}
	return b, err
}

// ListActiveByRoomBetween mengembalikan booking yang masih memakai ruang
// (booked, checked_in, completed) dan beririsan dengan rentang from-to.
func (r *BookingRepository) ListActiveByRoomBetween(ctx context.Context, roomID int64, from, to time.Time) ([]model.Booking, error) {
	return scanBookings(r.db.Query(ctx,
		`SELECT `+bookingColumns+` FROM bookings
		 WHERE room_id = $1 AND status IN ('booked', 'checked_in', 'completed')
		   AND start_at < $3 AND end_at > $2
		 ORDER BY start_at`, roomID, from, to))
}

// MarkCheckedIn mengubah booked menjadi checked_in. Kondisi status di WHERE
// mencegah check-in dobel bila ada dua request bersamaan.
func (r *BookingRepository) MarkCheckedIn(ctx context.Context, id int64, at time.Time) (model.Booking, error) {
	b, err := scanBooking(r.db.QueryRow(ctx,
		`UPDATE bookings SET status = 'checked_in', checked_in_at = $2, updated_at = NOW()
		 WHERE id = $1 AND status = 'booked'
		 RETURNING `+bookingColumns, id, at))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Booking{}, ErrBookingStatusChanged
	}
	return b, err
}

// MarkCheckedOut mengubah checked_in menjadi completed.
func (r *BookingRepository) MarkCheckedOut(ctx context.Context, id int64, at time.Time) (model.Booking, error) {
	b, err := scanBooking(r.db.QueryRow(ctx,
		`UPDATE bookings SET status = 'completed', checked_out_at = $2, updated_at = NOW()
		 WHERE id = $1 AND status = 'checked_in'
		 RETURNING `+bookingColumns, id, at))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Booking{}, ErrBookingStatusChanged
	}
	return b, err
}

// MarkNoShows mengubah semua booking yang belum check-in dan jam mulainya
// sebelum batas menjadi no_show, lalu mengembalikan booking yang diubah.
func (r *BookingRepository) MarkNoShows(ctx context.Context, startBefore time.Time) ([]model.Booking, error) {
	return scanBookings(r.db.Query(ctx,
		`UPDATE bookings SET status = 'no_show', updated_at = NOW()
		 WHERE status = 'booked' AND start_at < $1
		 RETURNING `+bookingColumns, startBefore))
}
