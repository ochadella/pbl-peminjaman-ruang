package model

import "time"

// Room adalah ruang diskusi perpustakaan yang bisa dipinjam.
type Room struct {
	ID         int64     `json:"id"`
	Code       string    `json:"code"`
	Name       string    `json:"name"`
	Location   string    `json:"location"`
	Capacity   int       `json:"capacity"`
	Facilities string    `json:"facilities"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CreateRoomRequest adalah body untuk POST /rooms (admin).
type CreateRoomRequest struct {
	Code       string `json:"code" validate:"required,max=20"`
	Name       string `json:"name" validate:"required,max=100"`
	Location   string `json:"location" validate:"max=100"`
	Capacity   int    `json:"capacity" validate:"required,gt=0"`
	Facilities string `json:"facilities"`
}

// UpdateRoomRequest adalah body untuk PATCH /rooms/:id (admin).
// Pointer dipakai agar field yang tidak dikirim tidak ikut diubah.
type UpdateRoomRequest struct {
	Name       *string `json:"name" validate:"omitempty,max=100"`
	Location   *string `json:"location" validate:"omitempty,max=100"`
	Capacity   *int    `json:"capacity" validate:"omitempty,gt=0"`
	Facilities *string `json:"facilities"`
	IsActive   *bool   `json:"is_active"`
}
