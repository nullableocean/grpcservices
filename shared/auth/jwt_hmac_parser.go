package auth

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrFailedParse  = errors.New("failed parse token")
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

type HmacJwtParser struct {
	key []byte
}

func NewHmacJwtParser(key string) *HmacJwtParser {
	return &HmacJwtParser{
		key: []byte(key),
	}
}

func (auth *HmacJwtParser) ParseToken(t string) (*UserClaims, error) {
	return auth.parseToken(t)
}

func (auth *HmacJwtParser) parseToken(t string) (*UserClaims, error) {
	claims := &UserClaims{}

	_, err := jwt.ParseWithClaims(t, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("expect HMAC signing method")
		}

		return auth.key, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}

		if errors.Is(err, jwt.ErrTokenInvalidClaims) {
			return nil, ErrInvalidToken
		}

		return nil, fmt.Errorf("%w: %w", ErrFailedParse, err)
	}

	return claims, nil
}
