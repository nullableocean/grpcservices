package auth

type JwtParser interface {
	ParseToken(t string) (*UserClaims, error)
}
