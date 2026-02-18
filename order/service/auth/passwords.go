package auth

import "golang.org/x/crypto/bcrypt"

type PasswordService struct{}

func (p *PasswordService) GetHashForPassword(pass string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func (p *PasswordService) CompareHashAndPassword(pass, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))

	return err == nil
}
