package service

import (
	"crypto/rand"
	"errors"
	"time"
)

const (
	
	MaxBookingDuration = 2 * time.Hour

	CancelDeadline = 30 * time.Minute

	checkinCodeLength   = 8
	checkinCodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" 
)

func ValidateBookingTime(start, end, now time.Time) error {
	switch {
	case !end.After(start):
		return errors.New("jam selesai harus setelah jam mulai")
	case end.Sub(start) > MaxBookingDuration:
		return errors.New("durasi booking maksimal 2 jam")
	case !start.After(now):
		return errors.New("jam mulai harus di masa depan")
	}
	return nil
}

func WithinOperatingHours(start, end, open, closeAt time.Time) bool {
	return !start.Before(open) && !end.After(closeAt)
}


func CanCancel(start, now time.Time) bool {
	return !now.After(start.Add(-CancelDeadline))
}

func NewCheckinCode() (string, error) {
	buf := make([]byte, checkinCodeLength)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	for i, b := range buf {
		buf[i] = checkinCodeAlphabet[int(b)%len(checkinCodeAlphabet)]
	}
	return string(buf), nil
}