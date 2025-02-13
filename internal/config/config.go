package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server       ServerConfig
	Database     DatabaseConfig
	TestDatabase TestDatabaseConfig
	JWT          JWTConfig
	OAuth        OAuthConfig
	SMTP         SMTPConfig
	SMS          SMSConfig
}

type ServerConfig struct {
	Port    int    `env:"SERVER_PORT"     default:"8080"`
	BaseURL string `env:"SERVER_BASE_URL" default:"http://localhost:8080"`
}

type DatabaseConfig struct {
	Host     string `env:"DB_HOST"     default:"localhost"`
	Port     int    `env:"DB_PORT"     default:"5432"`
	User     string `env:"DB_USER"     default:"postgres"`
	Password string `env:"DB_PASSWORD"                      required:"true"`
	Name     string `env:"DB_NAME"     default:"grocery_db"`
	SSLMode  string `env:"DB_SSLMODE"  default:"disable"`
}

type TestDatabaseConfig struct {
	Host     string `env:"TEST_DB_HOST"     default:"localhost"`
	Port     int    `env:"TEST_DB_PORT"     default:"5432"`
	User     string `env:"TEST_DB_USER"     default:"postgres"`
	Password string `env:"TEST_DB_PASSWORD"                      required:"true"`
	Name     string `env:"TEST_DB_NAME"     default:"grocery_db"`
	SSLMode  string `env:"TEST_DB_SSLMODE"  default:"disable"`
}

type JWTConfig struct {
	Secret        string        `env:"JWT_SECRET"         required:"true"`
	Issuer        string        `env:"JWT_ISSUER"                         default:"grocery-service"`
	TokenDuration time.Duration `env:"JWT_TOKEN_DURATION"                 default:"24h"`
}

type SMTPConfig struct {
	Host     string `env:"SMTP_HOST"      default:"smtp.resend.com"`
	Port     int    `env:"SMTP_PORT"      default:"465"`
	Username string `env:"SMTP_USERNAME"                            required:"true"`
	Password string `env:"SMTP_PASSWORD"                            required:"true"`
	From     string `env:"SMTP_FROM"                                required:"true"`
	FromName string `env:"SMTP_FROM_NAME" default:"Grocery Service"`
}

type SMSConfig struct {
	APIKey      string `env:"SMS_API_KEY"     required:"true"`
	Username    string `env:"SMS_USERNAME"    required:"true"`
	SenderID    string `env:"SMS_SENDER_ID"                   default:"GROCERY"`
	Environment string `env:"SMS_ENVIRONMENT"                 default:"sandbox"`
	BaseURL     string `env:"SMS_BASE_URL"`
}

type OAuthConfig struct {
	ClientID          string   `env:"OAUTH_CLIENT_ID"          required:"true"`
	ClientSecret      string   `env:"OAUTH_CLIENT_SECRET"      required:"true"`
	RedirectURL       string   `env:"OAUTH_REDIRECT_URL"       required:"true"`
	Scopes            []string `env:"OAUTH_SCOPES"`
	ProviderURL       string   `env:"OAUTH_PROVIDER_URL"       required:"true"`
	AuthorizeEndpoint string   `env:"OAUTH_AUTHORIZE_ENDPOINT"                 default:"/authorize"`
	TokenEndpoint     string   `env:"OAUTH_TOKEN_ENDPOINT"                     default:"/oauth/token"`
	UserInfoEndpoint  string   `env:"OAUTH_USERINFO_ENDPOINT"                  default:"/userinfo"`
}

func Load() (*Config, error) {
	cwd, _ := os.Getwd()
	fmt.Printf("Current working directory: %s\n", cwd)

	fmt.Println("Loading configuration from environment variables")

	config := &Config{
		Server: ServerConfig{
			Port: getEnvAsInt("SERVER_PORT", 8080),
			BaseURL: getEnv(
				"SERVER_BASE_URL",
				"http://localhost:8080",
			),
		},

		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "postges"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "grocery_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},

		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", ""),
			Issuer: getEnv("JWT_ISSUER", "grocery-service"),
			TokenDuration: getEnvAsDuration(
				"JWT_TOKEN_DURATION",
				24*time.Hour,
			),
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
			BaseURL: getEnv(
				"SMS_BASE_URL",
				"https://api.africastalking.com/version1/messaging",
			),
		},

		OAuth: OAuthConfig{
			ClientID:     getEnv("OAUTH_CLIENT_ID", ""),
			ClientSecret: getEnv("OAUTH_CLIENT_SECRET", ""),
			RedirectURL: getEnv(
				"OAUTH_REDIRECT_URL",
				"http://localhost:8080/api/v1/auth/callback",
			),
			Scopes: getEnvAsStringSlice(
				"OAUTH_SCOPES",
				[]string{
					"openid",
					"profile",
					"email",
				},
			),
			ProviderURL: getEnv(
				"OAUTH_PROVIDER_URL",
				"https://dev-vz8le2ezedv7udpb.us.auth0.com",
			),
			AuthorizeEndpoint: getEnv(
				"OAUTH_AUTHORIZE_ENDPOINT",
				"/v1/authorize",
			),
			TokenEndpoint: getEnv(
				"OAUTH_TOKEN_ENDPOINT",
				"/v1/token",
			),
		},
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func LoadTestConfig() (*TestDatabaseConfig, error) {
	defaultUser := "postgres"
	defaultPort := 5433
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		defaultUser = "runner"
		defaultPort = 5432
	}

	config := &TestDatabaseConfig{
		Host:     getEnv("TEST_DB_HOST", "localhost"),
		Port:     getEnvAsInt("TEST_DB_PORT", defaultPort),
		User:     getEnv("TEST_DB_USER", defaultUser),
		Password: getEnv("TEST_DB_PASSWORD", "postgres"),
		Name:     getEnv("TEST_DB_NAME", "grocery_test"),
		SSLMode:  getEnv("TEST_DB_SSLMODE", "disable"),
	}

	return config, nil
}

func (c *Config) Validate() error {
	var errors []string

	// Database validation
	if c.Database.Password == "" {
		errors = append(
			errors,
			"database password is required",
		)
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
		errors = append(
			errors,
			"SMTP from email is required",
		)
	}

	// SMS validation
	if c.SMS.APIKey == "" {
		errors = append(errors, "SMS API key is required")
	}

	if c.SMS.Username == "" {
		errors = append(errors, "SMS username is required")
	}

	// OAuth validation
	if c.OAuth.ClientID == "" {
		errors = append(
			errors,
			"OAuth client ID is required",
		)
	}

	if c.OAuth.ClientSecret == "" {
		errors = append(
			errors,
			"OAuth client secret is required",
		)
	}

	if c.OAuth.RedirectURL == "" {
		errors = append(
			errors,
			"OAuth redirect URL is required",
		)
	}

	if len(errors) > 0 {
		return fmt.Errorf(
			"configuration validation failed:\n- %s",
			strings.Join(errors, "\n- "),
		)
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists &&
		value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists &&
		value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(
	key string,
	defaultValue time.Duration,
) time.Duration {
	if value, exists := os.LookupEnv(key); exists &&
		value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvAsStringSlice(
	key string,
	defaultValue []string,
) []string {
	if value, exists := os.LookupEnv(key); exists &&
		value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
