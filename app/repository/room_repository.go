package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/ochadella/pbl-peminjaman-ruang/app/model"
)

var (
	ErrRoomNotFound  = errors.New("ruang tidak ditemukan")
	ErrRoomCodeTaken = errors.New("kode ruang sudah dipakai")
)

const roomColumns = `id, code, name, location, capacity, COALESCE(facilities, ''), is_active, created_at, updated_at`

type RoomRepository struct {
	db DBTX
}

func NewRoomRepository(db DBTX) *RoomRepository {
	return &RoomRepository{db: db}
}

func scanRoom(row pgx.Row) (model.Room, error) {
	var r model.Room
	err := row.Scan(&r.ID, &r.Code, &r.Name, &r.Location, &r.Capacity,
		&r.Facilities, &r.IsActive, &r.CreatedAt, &r.UpdatedAt)
	return r, err
}

// List mengembalikan semua ruang. Bila onlyActive true, ruang yang
// dinonaktifkan admin tidak ikut ditampilkan.
func (r *RoomRepository) List(ctx context.Context, onlyActive bool) ([]model.Room, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+roomColumns+` FROM rooms WHERE ($1 = FALSE OR is_active) ORDER BY code`, onlyActive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rooms := []model.Room{}
	for rows.Next() {
		room, err := scanRoom(rows)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, rows.Err()
}

func (r *RoomRepository) FindByID(ctx context.Context, id int64) (model.Room, error) {
	room, err := scanRoom(r.db.QueryRow(ctx, `SELECT `+roomColumns+` FROM rooms WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Room{}, ErrRoomNotFound
	}
	return room, err
}

func (r *RoomRepository) Create(ctx context.Context, req model.CreateRoomRequest) (model.Room, error) {
	room, err := scanRoom(r.db.QueryRow(ctx,
		`INSERT INTO rooms (code, name, location, capacity, facilities)
		 VALUES ($1, $2, $3, $4, NULLIF($5, ''))
		 RETURNING `+roomColumns,
		req.Code, req.Name, req.Location, req.Capacity, req.Facilities))
	if isUniqueViolation(err) {
		return model.Room{}, ErrRoomCodeTaken
	}
	return room, err
}

// Update hanya mengubah field yang dikirim (pointer tidak nil).
func (r *RoomRepository) Update(ctx context.Context, id int64, req model.UpdateRoomRequest) (model.Room, error) {
	room, err := scanRoom(r.db.QueryRow(ctx,
		`UPDATE rooms SET
		    name       = COALESCE($2, name),
		    location   = COALESCE($3, location),
		    capacity   = COALESCE($4, capacity),
		    facilities = COALESCE($5, facilities),
		    is_active  = COALESCE($6, is_active),
		    updated_at = NOW()
		 WHERE id = $1
		 RETURNING `+roomColumns,
		id, req.Name, req.Location, req.Capacity, req.Facilities, req.IsActive))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Room{}, ErrRoomNotFound
	}
	return room, err
}
