package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/ochadella/pbl-peminjaman-ruang/app/model"
)

var (
	ErrUserNotFound  = errors.New("user tidak ditemukan")
	ErrUserDuplicate = errors.New("email atau NIM/NIP sudah terdaftar")
)

const userColumns = `id, nim_nip, name, email, password_hash, role, is_active, blocked_until, created_at, updated_at`

type UserRepository struct {
	db DBTX
}

func NewUserRepository(db DBTX) *UserRepository {
	return &UserRepository{db: db}
}

func scanUser(row pgx.Row) (model.User, error) {
	var u model.User
	err := row.Scan(&u.ID, &u.NimNip, &u.Name, &u.Email, &u.PasswordHash,
		&u.Role, &u.IsActive, &u.BlockedUntil, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (r *UserRepository) Create(ctx context.Context, nimNip, name, email, passwordHash string) (model.User, error) {
	u, err := scanUser(r.db.QueryRow(ctx,
		`INSERT INTO users (nim_nip, name, email, password_hash, role)
		 VALUES ($1, $2, $3, $4, 'mahasiswa')
		 RETURNING `+userColumns,
		nimNip, name, email, passwordHash))
	if isUniqueViolation(err) {
		return model.User{}, ErrUserDuplicate
	}
	return u, err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (model.User, error) {
	u, err := scanUser(r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE email = $1`, email))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, ErrUserNotFound
	}
	return u, err
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (model.User, error) {
	u, err := scanUser(r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, ErrUserNotFound
	}
	return u, err
}

func (r *UserRepository) List(ctx context.Context) ([]model.User, error) {
	rows, err := r.db.Query(ctx, `SELECT `+userColumns+` FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []model.User{}
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *UserRepository) UpdateRole(ctx context.Context, id int64, role string) (model.User, error) {
	u, err := scanUser(r.db.QueryRow(ctx,
		`UPDATE users SET role = $2, updated_at = NOW() WHERE id = $1 RETURNING `+userColumns, id, role))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, ErrUserNotFound
	}
	return u, err
}
