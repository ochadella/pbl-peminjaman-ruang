package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/ochadella/pbl-peminjaman-ruang/app/model"
)

var (
	ErrOperatingHourNotFound = errors.New("jam operasional untuk hari ini belum diatur")
	ErrInvalidOperatingHour  = errors.New("jam tutup harus lebih besar dari jam buka")
)

// Kolom TIME diubah menjadi teks "HH:MM" agar mudah dibaca dan dibandingkan.
const operatingHourColumns = `id, day_of_week, to_char(open_time, 'HH24:MI'), to_char(close_time, 'HH24:MI'), is_open`

type OperatingHourRepository struct {
	db DBTX
}

func NewOperatingHourRepository(db DBTX) *OperatingHourRepository {
	return &OperatingHourRepository{db: db}
}

func scanOperatingHour(row pgx.Row) (model.OperatingHour, error) {
	var h model.OperatingHour
	err := row.Scan(&h.ID, &h.DayOfWeek, &h.OpenTime, &h.CloseTime, &h.IsOpen)
	return h, err
}

// List mengembalikan jam operasional seluruh hari, urut Minggu sampai Sabtu.
func (r *OperatingHourRepository) List(ctx context.Context) ([]model.OperatingHour, error) {
	rows, err := r.db.Query(ctx, `SELECT `+operatingHourColumns+` FROM operating_hours ORDER BY day_of_week`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hours := []model.OperatingHour{}
	for rows.Next() {
		h, err := scanOperatingHour(rows)
		if err != nil {
			return nil, err
		}
		hours = append(hours, h)
	}
	return hours, rows.Err()
}

// FindByDay dipakai saat mengecek ketersediaan ruang dan validasi booking.
func (r *OperatingHourRepository) FindByDay(ctx context.Context, day int) (model.OperatingHour, error) {
	h, err := scanOperatingHour(r.db.QueryRow(ctx,
		`SELECT `+operatingHourColumns+` FROM operating_hours WHERE day_of_week = $1`, day))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.OperatingHour{}, ErrOperatingHourNotFound
	}
	return h, err
}

// Upsert membuat atau mengganti jam operasional untuk satu hari.
func (r *OperatingHourRepository) Upsert(ctx context.Context, day int, req model.UpdateOperatingHourRequest) (model.OperatingHour, error) {
	h, err := scanOperatingHour(r.db.QueryRow(ctx,
		`INSERT INTO operating_hours (day_of_week, open_time, close_time, is_open)
		 VALUES ($1, $2::time, $3::time, $4)
		 ON CONFLICT (day_of_week) DO UPDATE
		   SET open_time = EXCLUDED.open_time,
		       close_time = EXCLUDED.close_time,
		       is_open = EXCLUDED.is_open
		 RETURNING `+operatingHourColumns,
		day, req.OpenTime, req.CloseTime, req.IsOpen))

	// 23514 = pelanggaran CHECK constraint (jam tutup <= jam buka).
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23514" {
		return model.OperatingHour{}, ErrInvalidOperatingHour
	}
	return h, err
}
