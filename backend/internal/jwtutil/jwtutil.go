package jwtutil

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	jwt.RegisteredClaims
	Roles []string `json:"roles,omitempty"`
}

func Sign(secret, issuer string, userID uuid.UUID, roles []string, ttl time.Duration) (string, error) {
	if secret == "" {
		return "", errors.New("jwt secret is empty")
	}
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        uuid.NewString(),
		},
		Roles: roles,
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	return tok.SignedString([]byte(secret))
}

func Parse(secret, issuer, tokenStr string) (*Claims, error) {
	tok, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil || !tok.Valid {
		return nil, ErrInvalidToken
	}
	claims, ok := tok.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}
	if issuer != "" && claims.Issuer != issuer {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func SubjectUUID(c *Claims) (uuid.UUID, error) {
	if c.Subject == "" {
		return uuid.Nil, ErrInvalidToken
	}
	return uuid.Parse(c.Subject)
}
