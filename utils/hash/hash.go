package hash

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Generate creates a hash from a password using bcrypt
func Generate(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %v", err)
	}
	return string(hashedBytes), nil
}

// Verify checks if the provided password matches the hashed password
func Verify(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
