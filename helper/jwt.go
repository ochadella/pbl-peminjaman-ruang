package helper

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenTTL adalah masa berlaku token login.
const TokenTTL = 24 * time.Hour

var ErrInvalidToken = errors.New("token tidak valid atau sudah kedaluwarsa")

// Claims adalah isi token: siapa user-nya dan apa role-nya.
type Claims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(secret string, userID int64, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(TokenTTL)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

func ParseToken(secret, tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims,
		func(t *jwt.Token) (any, error) { return []byte(secret), nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
