package auth

import "github.com/golang-jwt/jwt/v5"

type JwtAuthorizer interface {
	ParseToken(t string) (*jwt.Token, error)
	ExtractClaims(t *jwt.Token) (map[string]interface{}, error)
}
