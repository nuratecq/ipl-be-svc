package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for our application
type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Logger    LoggerConfig
	Doku      DokuConfig
	Mayar     MayarConfig
	JWT       JWTConfig
	CORS      CORSConfig
	Scheduler SchedulerConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port    string
	GinMode string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level  string
	Format string
}

// DokuConfig holds DOKU payment configuration (deprecated, use Mayar instead)
type DokuConfig struct {
	ClientID  string
	SecretKey string
	BaseURL   string
}

// MayarConfig holds Mayar payment configuration
type MayarConfig struct {
	AuthKey string
	BaseURL string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret string
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins string
}

// SchedulerConfig holds scheduler configuration
type SchedulerConfig struct {
	BillingCronExpression string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// It's okay if .env file doesn't exist
		fmt.Println("No .env file found, using environment variables")
	}

	config := &Config{
		Server: ServerConfig{
			Port:    getEnv("PORT", "8080"),
			GinMode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "192.168.8.187"),
			Port:     getEnvAsInt("DB_PORT", 54320),
			User:     getEnv("DB_USER", "admin"),
			Password: getEnv("DB_PASSWORD", "secret"),
			DBName:   getEnv("DB_NAME", "strapi"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Logger: LoggerConfig{
			Level:  getEnv("LOG_LEVEL", "debug"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		Doku: DokuConfig{
			ClientID:  getEnv("DOKU_CLIENT_ID", "BRN-0241-1762176502792"),
			SecretKey: getEnv("DOKU_SECRET_KEY", "SK-PaILsZudZTytTSTNCmUV"),
			BaseURL:   getEnv("DOKU_BASE_URL", "https://api-sandbox.doku.com"),
		},
		Mayar: MayarConfig{
			AuthKey: getEnv("MAYAR_AUTH_KEY", "your-mayar-auth-key"),
			BaseURL: getEnv("MAYAR_BASE_URL", "https://api.mayar.id/hl/v1"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "your-secret-key"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:3001,http://127.0.0.1:3000,http://127.0.0.1:3001"),
		},
		Scheduler: SchedulerConfig{
			BillingCronExpression: getEnv("BILLING_CRON_EXPRESSION", "0 0 0 1 * *"),
		},
	}

	return config, nil
}

// GetDSN returns PostgreSQL connection string
func (d *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode,
	)
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvAsInt gets an environment variable as integer with a fallback value
func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}
