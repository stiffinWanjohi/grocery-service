package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	SMTP     SMTPConfig
	SMS      SMSConfig
}

type ServerConfig struct {
	Port    int
	BaseURL string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	Secret string
	Issuer string
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
}

type SMSConfig struct {
	APIKey      string
	Username    string
	SenderID    string
	Environment string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: .env file not found: %v\n", err)
	}

	config := &Config{
		Server: ServerConfig{
			Port:    getEnvAsInt("SERVER_PORT", 8080),
			BaseURL: getEnv("SERVER_BASE_URL", "http://localhost:8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "grocery"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "your-secret-key"),
			Issuer: getEnv("JWT_ISSUER", "grocery-service"),
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			Port:     getEnvAsInt("SMTP_PORT", 587),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", "noreply@grocery.com"),
			FromName: getEnv("SMTP_FROM_NAME", "Grocery Service"),
		},
		SMS: SMSConfig{
			APIKey:      getEnv("SMS_API_KEY", ""),
			Username:    getEnv("SMS_USERNAME", ""),
			SenderID:    getEnv("SMS_SENDER_ID", "GROCERY"),
			Environment: getEnv("SMS_ENVIRONMENT", "sandbox"),
		},
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) Validate() error {
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}

	if c.Database.Password == "" {
		return fmt.Errorf("database password is required")
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
