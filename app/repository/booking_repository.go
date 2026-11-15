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
	ErrBookingConflict      = errors.New("jadwal booking bertabrakan")
	ErrDailyBookingExists   = errors.New("mahasiswa sudah menggunakan jatah booking hari ini")
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

// Create menyimpan booking baru milik mahasiswa.
func (r *BookingRepository) Create(
	ctx context.Context,
	userID, roomID int64,
	startAt, endAt time.Time,
	purpose string,
	participantCount int,
	checkinCode string,
) (model.Booking, error) {
	b, err := scanBooking(r.db.QueryRow(ctx,
		`INSERT INTO bookings (
			user_id, room_id, start_at, end_at,
			purpose, participant_count, status, checkin_code
		)
		VALUES ($1, $2, $3, $4, $5, $6, 'booked', $7)
		RETURNING `+bookingColumns,
		userID, roomID, startAt, endAt,
		purpose, participantCount, checkinCode,
	))

	if isUniqueViolation(err) {
		return model.Booking{}, ErrBookingConflict
	}

	return b, err
}

// ListByUser menampilkan seluruh booking milik satu mahasiswa.
func (r *BookingRepository) ListByUser(ctx context.Context, userID int64) ([]model.Booking, error) {
	return scanBookings(r.db.Query(ctx,
		`SELECT `+bookingColumns+`
		 FROM bookings
		 WHERE user_id = $1
		 ORDER BY start_at DESC`,
		userID,
	))
}

// Cancel membatalkan booking milik mahasiswa jika statusnya masih booked.
func (r *BookingRepository) Cancel(
	ctx context.Context,
	id, userID int64,
	at time.Time,
) (model.Booking, error) {
	b, err := scanBooking(r.db.QueryRow(ctx,
		`UPDATE bookings
		 SET status = 'cancelled',
		     cancelled_at = $3,
		     updated_at = NOW()
		 WHERE id = $1
		   AND user_id = $2
		   AND status = 'booked'
		 RETURNING `+bookingColumns,
		id, userID, at,
	))

	if errors.Is(err, pgx.ErrNoRows) {
		return model.Booking{}, ErrBookingStatusChanged
	}

	return b, err
}

// HasDailyBooking memeriksa apakah user sudah menggunakan jatah booking hari ini.
// Booking cancelled tidak dihitung, sedangkan no_show tetap dihitung.
func (r *BookingRepository) HasDailyBooking(
	ctx context.Context,
	userID int64,
	dayStart, dayEnd time.Time,
) (bool, error) {
	var exists bool

	err := r.db.QueryRow(ctx,
		`SELECT EXISTS (
            SELECT 1
            FROM bookings
            WHERE user_id = $1
              AND start_at >= $2
              AND start_at < $3
              AND status IN ('booked', 'checked_in', 'completed', 'no_show')
        )`,
		userID, dayStart, dayEnd,
	).Scan(&exists)

	return exists, err
}

// HasRoomConflict memeriksa apakah ada booking lain yang bertabrakan.
func (r *BookingRepository) HasRoomConflict(
	ctx context.Context,
	roomID int64,
	startAt, endAt time.Time,
) (bool, error) {
	var exists bool

	err := r.db.QueryRow(ctx,
		`SELECT EXISTS (
            SELECT 1
            FROM bookings
            WHERE room_id = $1
              AND status IN ('booked', 'checked_in', 'completed')
              AND start_at < $3
              AND end_at > $2
        )`,
		roomID, startAt, endAt,
	).Scan(&exists)

	return exists, err
}
