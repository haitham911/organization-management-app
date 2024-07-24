package utils

import (
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a plain text password.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash compares a plain text password with a hashed password.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
func RemoveLastSlash(url string) string {
	// Check if the URL ends with a slash
	if strings.HasSuffix(url, "/") {
		// Remove the last slash
		return strings.TrimSuffix(url, "/")
	}
	// Return the URL as is if it does not end with a slash
	return url
}
