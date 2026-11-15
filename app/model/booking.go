package model

import "time"

// Status booking. Alurnya: booked -> checked_in -> completed,
// atau booked -> cancelled / no_show.
const (
	BookingBooked    = "booked"
	BookingCheckedIn = "checked_in"
	BookingCompleted = "completed"
	BookingCancelled = "cancelled"
	BookingNoShow    = "no_show"
)

// Booking adalah peminjaman satu ruang oleh satu mahasiswa.
type Booking struct {
	ID               int64      `json:"id"`
	UserID           int64      `json:"user_id"`
	RoomID           int64      `json:"room_id"`
	StartAt          time.Time  `json:"start_at"`
	EndAt            time.Time  `json:"end_at"`
	Purpose          string     `json:"purpose"`
	ParticipantCount int        `json:"participant_count"`
	Status           string     `json:"status"`
	CheckinCode      *string    `json:"checkin_code"`
	CheckedInAt      *time.Time `json:"checked_in_at"`
	CheckedOutAt     *time.Time `json:"checked_out_at"`
	CancelledAt      *time.Time `json:"cancelled_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// CheckInRequest adalah body untuk POST /attendance/check-in.
// CreateBookingRequest adalah body untuk POST /bookings.
type CreateBookingRequest struct {
	RoomID           int64     `json:"room_id"`
	StartAt          time.Time `json:"start_at"`
	EndAt            time.Time `json:"end_at"`
	Purpose          string    `json:"purpose"`
	ParticipantCount int       `json:"participant_count"`
}

// CheckinCode adalah kode yang didapat dari hasil scan barcode.
type CheckInRequest struct {
	CheckinCode string `json:"checkin_code"`
}

// TimeSlot adalah rentang jam dalam format "HH:MM".
type TimeSlot struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// RoomAvailability adalah jadwal satu ruang pada satu tanggal.
type RoomAvailability struct {
	RoomID    int64      `json:"room_id"`
	Date      string     `json:"date"`
	IsOpen    bool       `json:"is_open"`
	OpenTime  string     `json:"open_time,omitempty"`
	CloseTime string     `json:"close_time,omitempty"`
	Booked    []TimeSlot `json:"booked"`
	Free      []TimeSlot `json:"free"`
}
