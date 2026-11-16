package model

import "time"

type CreateBookingRequest struct {
	RoomID           int64     `json:"room_id"`
	StartAt          time.Time `json:"start_at"`
	EndAt            time.Time `json:"end_at"`
	Purpose          string    `json:"purpose"`
	ParticipantCount int       `json:"participant_count"`
}