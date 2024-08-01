package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	Protocol    string
	secretKey   string
	rabbitMqUri string
	ContentType string
}

// LoadConfig reads configuration from environment variables or .env file
func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables if set.")
	}

	return &Config{
		Port:        getEnv("PORT", "8080"),
		Protocol:    getEnv("PROTOCOL", "tcp"),
		ContentType: getEnv("CONTENT_TYPE", "application/json"),
		secretKey:   getEnv("SECRET_KEY", "prodonik"),
		rabbitMqUri: getEnv("RABBITMQ_URI", "amqp://rabbitmq:5672"),
	}, nil
}

// Helper function to get environment variables with a fallback value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func (c *Config) GetSecretKey() string {
	return c.secretKey
}

func (c *Config) GetRabbitMqURI() string {
	return c.rabbitMqUri
}
