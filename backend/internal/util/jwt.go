package util

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func CreateToken(id uint, role, secret, issuer string) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{Role: role, RegisteredClaims: jwt.RegisteredClaims{Subject: fmt.Sprint(id), Issuer: issuer, ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour))}}).SignedString([]byte(secret))
}
func ParseToken(raw, secret, issuer string) (*Claims, error) {
	token, e := jwt.ParseWithClaims(raw, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	}, jwt.WithIssuer(issuer))
	if e != nil {
		return nil, e
	}
	c, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return c, nil
}
