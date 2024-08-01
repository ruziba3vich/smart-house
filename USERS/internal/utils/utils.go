package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/ruziba3vich/users/internal/config"
)

type (
	TokenGenerator struct {
		secretKey string
	}
	PasswordHasher struct{}
)

// NewTokenGenerator creates a new TokenGenerator
func NewTokenGenerator(cfg *config.Config) *TokenGenerator {
	return &TokenGenerator{
		secretKey: cfg.GetSecretKey(),
	}
}

// GenerateToken generates a JWT token for a user
func (t *TokenGenerator) GenerateToken(userId string, username string) (string, error) {
	claims := jwt.MapClaims{
		"sub":      userId,
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(t.secretKey))
	if err != nil {
		return "", fmt.Errorf("could not create token: %s", err.Error())
	}

	return tokenString, nil
}

// HashPassword hashes a password using SHA-256
func (p *PasswordHasher) HashPassword(password string) string {
	hasher := sha256.New()
	hasher.Write([]byte(password))
	return hex.EncodeToString(hasher.Sum(nil))
}

// CheckPasswordHash compares a plain password with its hash
func (p *PasswordHasher) CheckPasswordHash(password, hash string) bool {
	return p.HashPassword(password) == hash
}

// NewPasswordHasher creates a new PasswordHasher
func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{}
}
