package model

import "time"

const (
	RoleMahasiswa = "mahasiswa"
	RolePetugas   = "petugas"
	RoleAdmin     = "admin"
)

// User adalah akun sistem. PasswordHash tidak pernah dikirim ke client.
type User struct {
	ID           int64      `json:"id"`
	NimNip       string     `json:"nim_nip"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Role         string     `json:"role"`
	IsActive     bool       `json:"is_active"`
	BlockedUntil *time.Time `json:"blocked_until"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type RegisterRequest struct {
	NimNip   string `json:"nim_nip"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type UpdateRoleRequest struct {
	Role string `json:"role"`
}
