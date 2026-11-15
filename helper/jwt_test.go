package helper

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "secret-untuk-testing-minimal-32-karakter"

func TestGenerateDanParseToken(t *testing.T) {
	token, err := GenerateToken(testSecret, 7, "petugas")
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	claims, err := ParseToken(testSecret, token)
	if err != nil {
		t.Fatalf("ParseToken error: %v", err)
	}
	if claims.UserID != 7 {
		t.Errorf("UserID = %d, seharusnya 7", claims.UserID)
	}
	if claims.Role != "petugas" {
		t.Errorf("Role = %q, seharusnya petugas", claims.Role)
	}
	if got := claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time); got != TokenTTL {
		t.Errorf("masa berlaku = %v, seharusnya %v", got, TokenTTL)
	}
}

func TestParseToken_Gagal(t *testing.T) {
	valid, _ := GenerateToken(testSecret, 1, "mahasiswa")

	expired, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: 1, Role: "mahasiswa",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-48 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-24 * time.Hour)),
		},
	}).SignedString([]byte(testSecret))

	metodeLain, _ := jwt.NewWithClaims(jwt.SigningMethodHS512, Claims{
		UserID: 1, Role: "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}).SignedString([]byte(testSecret))

	tests := []struct {
		name   string
		secret string
		token  string
	}{
		{"secret salah", "secret-yang-lain-sama-sekali-beda", valid},
		{"token kedaluwarsa", testSecret, expired},
		{"metode signing bukan HS256", testSecret, metodeLain},
		{"token ngawur", testSecret, "bukan-token-jwt"},
		{"token kosong", testSecret, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ParseToken(tt.secret, tt.token)
			if !errors.Is(err, ErrInvalidToken) {
				t.Errorf("err = %v, seharusnya ErrInvalidToken", err)
			}
			if claims != nil {
				t.Errorf("claims seharusnya nil, dapat %+v", claims)
			}
		})
	}
}