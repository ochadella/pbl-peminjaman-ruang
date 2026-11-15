package service

import (
	"testing"
	"time"
)

func TestCanCheckIn(t *testing.T) {
	start := time.Date(2026, 11, 20, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		now  time.Time
		want bool
	}{
		{"terlalu awal (16 menit sebelum)", start.Add(-16 * time.Minute), false},
		{"tepat 15 menit sebelum", start.Add(-15 * time.Minute), true},
		{"tepat jam mulai", start, true},
		{"tepat 15 menit sesudah", start.Add(15 * time.Minute), true},
		{"terlambat (16 menit sesudah)", start.Add(16 * time.Minute), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanCheckIn(start, tt.now); got != tt.want {
				t.Errorf("CanCheckIn() = %v, seharusnya %v", got, tt.want)
			}
		})
	}
}

func TestIsNoShow(t *testing.T) {
	start := time.Date(2026, 11, 20, 10, 0, 0, 0, time.UTC)
	if IsNoShow(start, start.Add(15*time.Minute)) {
		t.Error("tepat 15 menit sesudah mulai belum boleh dianggap no_show")
	}
	if !IsNoShow(start, start.Add(16*time.Minute)) {
		t.Error("16 menit sesudah mulai seharusnya sudah no_show")
	}
}

func TestBlockedUntil(t *testing.T) {
	now := time.Date(2026, 11, 20, 10, 0, 0, 0, time.UTC)
	if got := BlockedUntil(2, now); got != nil {
		t.Errorf("2 pelanggaran belum boleh diblokir, dapat %v", got)
	}
	got := BlockedUntil(3, now)
	if got == nil {
		t.Fatal("3 pelanggaran seharusnya diblokir")
	}
	if want := now.AddDate(0, 0, 7); !got.Equal(want) {
		t.Errorf("blokir sampai %v, seharusnya %v", got, want)
	}
}

func TestIsLateCheckout(t *testing.T) {
	end := time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		now  time.Time
		want bool
	}{
		{"sebelum jam selesai", end.Add(-10 * time.Minute), false},
		{"tepat 15 menit sesudah", end.Add(15 * time.Minute), false},
		{"16 menit sesudah", end.Add(16 * time.Minute), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsLateCheckout(end, tt.now); got != tt.want {
				t.Errorf("IsLateCheckout() = %v, want %v", got, tt.want)
			}
		})
	}
}
