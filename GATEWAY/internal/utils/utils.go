package utils

import (
	"encoding/hex"
	"fmt"
	"hash"

	"github.com/golang-jwt/jwt"
	"github.com/ruziba3vich/smart-house/internal/config"
)

type (
	PasswordHasher struct {
		hash hash.Hash
	}

	TokenGenerator struct {
		secretKey string
	}
)

func NewPasswordHasher(hash hash.Hash) *PasswordHasher {
	return &PasswordHasher{
		hash: hash,
	}
}

func NewTokenGenerator(cfg *config.Config) *TokenGenerator {
	return &TokenGenerator{
		secretKey: cfg.GetSecretKey(),
	}
}

func (p *PasswordHasher) HashPassword(password string) string {
	p.hash.Write([]byte(password))
	return hex.EncodeToString(p.hash.Sum(nil))
}

func (t *TokenGenerator) ExtractUserData(tokenString string) (string, string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(t.secretKey), nil
	})

	if err != nil {
		return "", "", fmt.Errorf("could not parse token: %s", err.Error())
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userId, ok1 := claims["sub"].(string)
		username, ok2 := claims["username"].(string)
		if !ok1 || !ok2 {
			return "", "", fmt.Errorf("could not extract user data from token")
		}
		return userId, username, nil
	}

	return "", "", fmt.Errorf("invalid token")
}
