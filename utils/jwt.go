package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims is the payload stored inside the token.
type JWTClaims struct {
	UserID       uint   `json:"user_id"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	IsSuperAdmin bool   `json:"is_super_admin"`
	// TokenVersion must match the user's CURRENT token_version in the DB
	// at request time — every login bumps that DB value, so a token
	// issued by an earlier login carries a stale version and gets
	// rejected. This is what enforces "one active session per user":
	// logging in on a new device invalidates every token from previous
	// logins without needing a server-side session store or token
	// blocklist.
	TokenVersion int `json:"token_version"`
	jwt.RegisteredClaims
}

// GenerateToken creates a signed JWT string.
func GenerateToken(userID uint, email, role, secret string, expireHours int, isSuperAdmin bool, tokenVersion int) (string, error) {
	claims := JWTClaims{
		UserID:       userID,
		Email:        email,
		Role:         role,
		IsSuperAdmin: isSuperAdmin,
		TokenVersion: tokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

// ParseToken validates the token and returns claims.
func ParseToken(tokenStr, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
