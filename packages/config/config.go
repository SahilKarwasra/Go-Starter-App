package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                 string
	DatabaseUrl          string
	JwtSecret            string
	JwtRefreshSecret     string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

func LoadConfig() (*Config, error) {
	// Search for .env from current working directory up to 3 levels
	_ = godotenv.Load(".env", "../.env", "../../.env", "../../../.env")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is not set")
	}

	jwtRefreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	if jwtRefreshSecret == "" {
		jwtRefreshSecret = jwtSecret + "_refresh"
	}

	accessDuration := 15 * time.Minute
	if d := os.Getenv("ACCESS_TOKEN_EXPIRY"); d != "" {
		if parsed, err := time.ParseDuration(d); err == nil {
			accessDuration = parsed
		}
	}

	refreshDuration := 7 * 24 * time.Hour
	if d := os.Getenv("REFRESH_TOKEN_EXPIRY"); d != "" {
		if parsed, err := time.ParseDuration(d); err == nil {
			refreshDuration = parsed
		}
	}

	return &Config{
		Port:                 port,
		DatabaseUrl:          dbURL,
		JwtSecret:            jwtSecret,
		JwtRefreshSecret:     jwtRefreshSecret,
		AccessTokenDuration:  accessDuration,
		RefreshTokenDuration: refreshDuration,
	}, nil
}
