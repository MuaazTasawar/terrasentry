package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	Env             string
	DatabaseURL     string
	RiskScoringURL  string
	JWTSecret       string
	FCMServerKey    string
	KubeconfigPath  string
}

func Load() *Config {
	// .env is optional in production (real env vars take over), but load it if present
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on system environment variables")
	}

	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		Env:            getEnv("ENV", "development"),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		RiskScoringURL: getEnv("RISK_SCORING_URL", "http://localhost:8000"),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		FCMServerKey:   getEnv("FCM_SERVER_KEY", ""),
		KubeconfigPath: getEnv("KUBECONFIG_PATH", ""),
	}

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required but not set")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}