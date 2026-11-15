package model

import "time"

// Jenis pelanggaran yang dikenal sistem.
const (
	ViolationNoShow       = "no_show"
	ViolationLateCheckout = "telat_checkout"
	ViolationDamage       = "kerusakan"
)

// Batas pelanggaran sebelum mahasiswa diblokir booking.
const (
	MaxViolations    = 3
	BlockDurationDay = 7
)

// Violation adalah catatan pelanggaran seorang mahasiswa.
// BookingID kosong (nil) bila pelanggaran tidak terkait booking tertentu.
// RecordedBy kosong (nil) bila dicatat otomatis oleh sistem.
type Violation struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	BookingID  *int64    `json:"booking_id"`
	Type       string    `json:"type"`
	Note       string    `json:"note"`
	RecordedBy *int64    `json:"recorded_by"`
	CreatedAt  time.Time `json:"created_at"`
}

// CreateViolationRequest adalah body untuk POST /violations (petugas).
type CreateViolationRequest struct {
	UserID    int64  `json:"user_id" validate:"required,gt=0"`
	BookingID *int64 `json:"booking_id"`
	Type      string `json:"type" validate:"required,oneof=no_show telat_checkout kerusakan"`
	Note      string `json:"note" validate:"max=500"`
}
