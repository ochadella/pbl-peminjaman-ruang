package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/ochadella/pbl-peminjaman-ruang/app/model"
)

var ErrViolationExists = errors.New("booking ini sudah tercatat sebagai pelanggaran")

const violationColumns = `id, user_id, booking_id, type, COALESCE(note, ''), recorded_by, created_at`

type ViolationRepository struct {
	db DBTX
}

func NewViolationRepository(db DBTX) *ViolationRepository {
	return &ViolationRepository{db: db}
}

func scanViolation(row pgx.Row) (model.Violation, error) {
	var v model.Violation
	err := row.Scan(&v.ID, &v.UserID, &v.BookingID, &v.Type, &v.Note, &v.RecordedBy, &v.CreatedAt)
	return v, err
}

// Create mencatat pelanggaran. recordedBy nil berarti dicatat otomatis oleh sistem.
func (r *ViolationRepository) Create(ctx context.Context, req model.CreateViolationRequest, recordedBy *int64) (model.Violation, error) {
	v, err := scanViolation(r.db.QueryRow(ctx,
		`INSERT INTO violations (user_id, booking_id, type, note, recorded_by)
		 VALUES ($1, $2, $3, NULLIF($4, ''), $5)
		 RETURNING `+violationColumns,
		req.UserID, req.BookingID, req.Type, req.Note, recordedBy))
	if isUniqueViolation(err) {
		return model.Violation{}, ErrViolationExists
	}
	return v, err
}

// ListByUser menampilkan riwayat pelanggaran satu mahasiswa, terbaru lebih dulu.
func (r *ViolationRepository) ListByUser(ctx context.Context, userID int64) ([]model.Violation, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+violationColumns+` FROM violations WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []model.Violation{}
	for rows.Next() {
		v, err := scanViolation(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, v)
	}
	return list, rows.Err()
}

// CountByUser dipakai untuk menentukan apakah mahasiswa perlu diblokir.
func (r *ViolationRepository) CountByUser(ctx context.Context, userID int64) (int, error) {
	var n int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM violations WHERE user_id = $1`, userID).Scan(&n)
	return n, err
}

// SetBlockedUntil mengisi kolom users.blocked_until saat batas pelanggaran tercapai.
func (r *ViolationRepository) SetBlockedUntil(ctx context.Context, userID int64, until time.Time) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET blocked_until = $2, updated_at = NOW() WHERE id = $1`, userID, until)
	return err
}
