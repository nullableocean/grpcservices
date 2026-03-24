package auth

import (
	"errors"
	"maps"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

type HmacJwtAuth struct {
	key []byte
}

func NewHmacJwtAuth(key string) *HmacJwtAuth {
	return &HmacJwtAuth{
		key: []byte(key),
	}
}

func (auth *HmacJwtAuth) ParseToken(t string) (*jwt.Token, error) {
	return auth.parseToken(t)
}

func (auth *HmacJwtAuth) ExtractClaims(t *jwt.Token) (map[string]interface{}, error) {
	return auth.getTokenClaims(t)
}

func (auth *HmacJwtAuth) parseToken(t string) (*jwt.Token, error) {
	token, err := jwt.Parse(t, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("expect HMAC signing method")
		}

		return auth.key, nil
	})

	if err != nil {
		return nil, errors.Join(ErrInvalidToken, err)
	}

	return token, nil
}

func (auth *HmacJwtAuth) getTokenClaims(token *jwt.Token) (map[string]interface{}, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	outClaims := make(map[string]interface{}, len(claims))
	maps.Copy(outClaims, claims)

	return outClaims, nil
}
