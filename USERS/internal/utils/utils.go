package utils

import (
	"encoding/hex"
	"fmt"
	"hash"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/ruziba3vich/users/internal/config"
)

type PasswordHasher struct {
	hash hash.Hash
}

// / HashPassword hashes a password using SHA-256
func (p *PasswordHasher) HashPassword(password string) string {
	p.hash.Write([]byte(password))
	return hex.EncodeToString(p.hash.Sum(nil))
}

func NewPasswordHasher(hash hash.Hash) *PasswordHasher {
	return &PasswordHasher{
		hash: hash,
	}
}

type TokenGenerator struct {
	secretKey string
}

func NewTokenGenerator(cfg *config.Config) *TokenGenerator {
	return &TokenGenerator{
		secretKey: cfg.GetSecretKey(),
	}
}

// / method to generate a new token
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
