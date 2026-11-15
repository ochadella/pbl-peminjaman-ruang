package service

import (
	"time"

	"github.com/ochadella/pbl-peminjaman-ruang/app/model"
)

// CheckInWindow adalah jarak waktu sebelum dan sesudah jam mulai booking
// di mana mahasiswa masih boleh check-in.
const CheckInWindow = 15 * time.Minute

// CanCheckIn memeriksa apakah waktu now masih berada di jendela check-in:
// paling cepat 15 menit sebelum jam mulai, paling lambat 15 menit sesudahnya.
func CanCheckIn(startAt, now time.Time) bool {
	return !now.Before(startAt.Add(-CheckInWindow)) && !now.After(startAt.Add(CheckInWindow))
}

// IsNoShow bernilai true bila booking belum check-in dan sudah lewat
// 15 menit dari jam mulai, sehingga booking harus diubah menjadi no_show.
func IsNoShow(startAt, now time.Time) bool {
	return now.After(startAt.Add(CheckInWindow))
}

// BlockedUntil menghitung sampai kapan mahasiswa diblokir booking.
// Mengembalikan nil bila jumlah pelanggaran belum mencapai batas.
func BlockedUntil(violationCount int, now time.Time) *time.Time {
	if violationCount < model.MaxViolations {
		return nil
	}
	until := now.AddDate(0, 0, model.BlockDurationDay)
	return &until
}

// LateCheckoutTolerance adalah batas toleransi check-out setelah jam selesai.
const LateCheckoutTolerance = 15 * time.Minute

// IsLateCheckout bernilai true bila check-out dilakukan lebih dari
// 15 menit setelah jam selesai booking.
func IsLateCheckout(endAt, now time.Time) bool {
	return now.After(endAt.Add(LateCheckoutTolerance))
}
