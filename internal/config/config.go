package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

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
	Port    int    `env:"SERVER_PORT" default:"8080"`
	BaseURL string `env:"SERVER_BASE_URL" default:"http://localhost:8080"`
}

type DatabaseConfig struct {
	Host     string `env:"DB_HOST" default:"localhost"`
	Port     int    `env:"DB_PORT" default:"5432"`
	User     string `env:"DB_USER" default:"postgres"`
	Password string `env:"DB_PASSWORD" required:"true"`
	Name     string `env:"DB_NAME" default:"grocery_db"`
	SSLMode  string `env:"DB_SSLMODE" default:"disable"`
}

type JWTConfig struct {
	Secret string `env:"JWT_SECRET" required:"true"`
	Issuer string `env:"JWT_ISSUER" default:"grocery-service"`
}

type SMTPConfig struct {
	Host     string `env:"SMTP_HOST" default:"smtp.gmail.com"`
	Port     int    `env:"SMTP_PORT" default:"587"`
	Username string `env:"SMTP_USERNAME" required:"true"`
	Password string `env:"SMTP_PASSWORD" required:"true"`
	From     string `env:"SMTP_FROM" required:"true"`
	FromName string `env:"SMTP_FROM_NAME" default:"Grocery Service"`
}

type SMSConfig struct {
	APIKey      string `env:"SMS_API_KEY" required:"true"`
	Username    string `env:"SMS_USERNAME" required:"true"`
	SenderID    string `env:"SMS_SENDER_ID" default:"GROCERY"`
	Environment string `env:"SMS_ENVIRONMENT" default:"sandbox"`
	BaseURL     string `env:"SMS_BASE_URL" default:"https://api.africastalking.com/version1/messaging"`
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("error loading .env file: %w", err)
		}
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
			Name:     getEnv("DB_NAME", "grocery_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", ""),
			Issuer: getEnv("JWT_ISSUER", "grocery-service"),
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			Port:     getEnvAsInt("SMTP_PORT", 587),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", ""),
			FromName: getEnv("SMTP_FROM_NAME", "Grocery Service"),
		},
		SMS: SMSConfig{
			APIKey:      getEnv("SMS_API_KEY", ""),
			Username:    getEnv("SMS_USERNAME", ""),
			SenderID:    getEnv("SMS_SENDER_ID", "GROCERY"),
			Environment: getEnv("SMS_ENVIRONMENT", "sandbox"),
			BaseURL:     getEnv("SMS_BASE_URL", "https://api.africastalking.com/version1/messaging"),
		},
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) Validate() error {
	var errors []string

	// Database validation
	if c.Database.Password == "" {
		errors = append(errors, "database password is required")
	}

	// JWT validation
	if c.JWT.Secret == "" {
		errors = append(errors, "JWT secret is required")
	}

	// SMTP validation
	if c.SMTP.Username == "" {
		errors = append(errors, "SMTP username is required")
	}
	if c.SMTP.Password == "" {
		errors = append(errors, "SMTP password is required")
	}
	if c.SMTP.From == "" {
		errors = append(errors, "SMTP from email is required")
	}

	// SMS validation
	if c.SMS.APIKey == "" {
		errors = append(errors, "SMS API key is required")
	}
	if c.SMS.Username == "" {
		errors = append(errors, "SMS username is required")
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n- %s", strings.Join(errors, "\n- "))
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
