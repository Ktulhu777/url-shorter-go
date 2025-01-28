package hash_password

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const totalIterations int = 14

func GeneratePassword(password string) (string, error) {
	const fn = "lib.hash_password.generatePassword"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), totalIterations)
	if err != nil {
		return "", fmt.Errorf("%s: failed to hash password: %w", fn, err)
	}

	return string(hash), nil
}
