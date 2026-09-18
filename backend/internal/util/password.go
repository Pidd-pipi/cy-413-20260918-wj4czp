package util

import "golang.org/x/crypto/bcrypt"

func HashPassword(s string) (string, error) {
	b, e := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
	return string(b), e
}
func ComparePassword(hash, s string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(s))
}
