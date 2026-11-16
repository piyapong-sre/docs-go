package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// MaxPasswordBytes is bcrypt's input limit. GenerateFromPassword returns
// bcrypt.ErrPasswordTooLong for longer passwords, so validate this first.
const MaxPasswordBytes = 72

// HashPassword hashes a plaintext password with bcrypt.
func HashPassword(plain string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

// CheckPassword reports whether plain matches hash.
// It returns (false, nil) for a wrong password and an error only when
// the hash itself is unusable.
func CheckPassword(hash, plain string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
		return false, nil
	default:
		return false, err
	}
}
