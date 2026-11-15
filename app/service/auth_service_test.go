package service

import (
	"strings"
	"testing"

	"github.com/ochadella/pbl-peminjaman-ruang/app/model"
)

func TestValidateRegister(t *testing.T) {
	roem := model.RegisterRequest{
		NimNip:   "2301010001",
		Name:     "Roem Hanan",
		Email:    "roem.hanan@student.ac.id",
		Password: "PasswordRoem123",
	}

	ubah := func(f func(r *model.RegisterRequest)) model.RegisterRequest {
		r := roem
		f(&r)
		return r
	}

	tests := []struct {
		name    string
		req     model.RegisterRequest
		wantErr string 
	}{
		{"data Roem Hanan valid", roem, ""},
		{"Kevin Rio valid", model.RegisterRequest{NimNip: "2301010002", Name: "Kevin Rio",
			Email: "kevin.rio@student.ac.id", Password: "PasswordKevin456"}, ""},
		{"password tepat 8 karakter", ubah(func(r *model.RegisterRequest) { r.Password = "12345678" }), ""},
		{"NIM tepat 20 karakter", ubah(func(r *model.RegisterRequest) { r.NimNip = strings.Repeat("1", 20) }), ""},

		{"NIM kosong", ubah(func(r *model.RegisterRequest) { r.NimNip = "" }), "NIM/NIP"},
		{"NIM hanya spasi", ubah(func(r *model.RegisterRequest) { r.NimNip = "   " }), "NIM/NIP"},
		{"NIM 21 karakter", ubah(func(r *model.RegisterRequest) { r.NimNip = strings.Repeat("1", 21) }), "NIM/NIP"},
		{"nama kosong", ubah(func(r *model.RegisterRequest) { r.Name = "" }), "nama"},
		{"nama hanya spasi", ubah(func(r *model.RegisterRequest) { r.Name = "   " }), "nama"},
		{"email tanpa @", ubah(func(r *model.RegisterRequest) { r.Email = "asilla.cesa.student.ac.id" }), "email"},
		{"email kosong", ubah(func(r *model.RegisterRequest) { r.Email = "" }), "email"},
		{"password 7 karakter", ubah(func(r *model.RegisterRequest) { r.Password = "1234567" }), "password"},
		{"password kosong", ubah(func(r *model.RegisterRequest) { r.Password = "" }), "password"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRegister(tt.req)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("seharusnya lolos, dapat error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("seharusnya error mengandung %q, tapi lolos", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, seharusnya mengandung %q", err.Error(), tt.wantErr)
			}
		})
	}
}