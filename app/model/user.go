package model

import "time"

type User struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	NimNip       string     `gorm:"column:nim_nip;uniqueIndex" json:"nim_nip"`
	Name         string     `json:"name"`
	Email        string     `gorm:"uniqueIndex" json:"email"`
	PasswordHash string     `json:"-"`
	Role         string     `json:"role"`
	IsActive     bool       `json:"is_active"`
	BlockedUntil *time.Time `json:"blocked_until"`
}

const (
	RoleMahasiswa = "mahasiswa"
	RolePetugas   = "petugas"
	RoleAdmin     = "admin"
)