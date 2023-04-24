package utils

import (
	"ArchGo/logger"
	"crypto/rand"
	"encoding/base32"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", fmt.Errorf("could not hash password %w", err)
	}
	return string(hashedPassword), nil
}

func VerifyPassword(hashedPassword string, candidatePassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(candidatePassword))
}

func GenerateRandomString(length int) string {
	randomBytes := make([]byte, (length*5+7)/8) // Calculate the number of bytes required for the desired length
	_, err := rand.Read(randomBytes)
	if err != nil {
		logger.Warning("Error while generating random bytes for prompt %s", err.Error())
		return ""
	}

	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	return encoded[:length]
}
