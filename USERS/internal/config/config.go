package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// DbConfig holds the database configuration
type DbConfig struct {
	MongoURI   string
	MongoDB    string
	Collection string
}

// Config holds the application configuration
type Config struct {
	DbConfig  DbConfig
	Port      string
	secretKey string
}

// LoadConfig reads configuration from environment variables or .env file
func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables if set.")
	}

	return &Config{
		DbConfig: DbConfig{
			MongoURI:   getEnv("MONGO_URI", "mongodb://localhost:27017"),
			MongoDB:    getEnv("MONGO_DB", "test"),
			Collection: getEnv("MONGO_COLLECTION", "users"),
		},
		Port:      getEnv("PORT", "8080"),
		secretKey: getEnv("SECRET_KEY", "prodonik"),
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
