package service

import (
	"strings"
	"testing"
	"time"
)

func TestValidateBookingTime(t *testing.T) {
	now := time.Date(2026, 11, 19, 9, 0, 0, 0, WIB)
	start := time.Date(2026, 11, 20, 10, 0, 0, 0, WIB)

	tests := []struct {
		name    string
		start   time.Time
		end     time.Time
		wantErr string 
	}{
		{"durasi 1 jam", start, start.Add(time.Hour), ""},
		{"durasi tepat 2 jam", start, start.Add(2 * time.Hour), ""},
		{"durasi 2 jam 1 menit", start, start.Add(2*time.Hour + time.Minute), "maksimal 2 jam"},
		{"jam selesai sama dengan mulai", start, start, "setelah jam mulai"},
		{"jam selesai sebelum mulai", start, start.Add(-time.Hour), "setelah jam mulai"},
		{"jam mulai sudah lewat", now.Add(-time.Hour), now, "masa depan"},
		{"jam mulai tepat sekarang", now, now.Add(time.Hour), "masa depan"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBookingTime(tt.start, tt.end, now)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("seharusnya lolos, dapat error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %v, seharusnya mengandung %q", err, tt.wantErr)
			}
		})
	}
}

func TestWithinOperatingHours(t *testing.T) {
	open := time.Date(2026, 11, 20, 8, 0, 0, 0, WIB)
	closeAt := time.Date(2026, 11, 20, 16, 0, 0, 0, WIB)
	at := func(h, m int) time.Time { return time.Date(2026, 11, 20, h, m, 0, 0, WIB) }

	tests := []struct {
		name       string
		start, end time.Time
		want       bool
	}{
		{"di tengah jam buka", at(10, 0), at(12, 0), true},
		{"mulai tepat jam buka", at(8, 0), at(10, 0), true},
		{"selesai tepat jam tutup", at(14, 0), at(16, 0), true},
		{"mulai sebelum jam buka", at(7, 30), at(9, 30), false},
		{"selesai setelah jam tutup", at(15, 0), at(16, 30), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithinOperatingHours(tt.start, tt.end, open, closeAt); got != tt.want {
				t.Errorf("WithinOperatingHours() = %v, seharusnya %v", got, tt.want)
			}
		})
	}
}

func TestCanCancel(t *testing.T) {
	start := time.Date(2026, 11, 20, 10, 0, 0, 0, WIB)
	tests := []struct {
		name string
		now  time.Time
		want bool
	}{
		{"sehari sebelumnya", start.Add(-24 * time.Hour), true},
		{"tepat 30 menit sebelum", start.Add(-30 * time.Minute), true},
		{"29 menit sebelum", start.Add(-29 * time.Minute), false},
		{"setelah jam mulai", start.Add(time.Minute), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanCancel(start, tt.now); got != tt.want {
				t.Errorf("CanCancel() = %v, seharusnya %v", got, tt.want)
			}
		})
	}
}

func TestNewCheckinCode(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		code, err := NewCheckinCode()
		if err != nil {
			t.Fatalf("NewCheckinCode error: %v", err)
		}
		if len(code) != checkinCodeLength {
			t.Fatalf("panjang kode = %d, seharusnya %d", len(code), checkinCodeLength)
		}
		for _, r := range code {
			if !strings.ContainsRune(checkinCodeAlphabet, r) {
				t.Fatalf("kode %q memuat karakter %q di luar alfabet", code, r)
			}
		}
		seen[code] = true
	}
	if len(seen) < 49 {
		t.Errorf("50 kode hanya menghasilkan %d kode berbeda, terlalu banyak yang kembar", len(seen))
	}
}