package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBSource  string
	Port      string
	JWTSecret string
	AppEnv    string
}

func Load() (*Config, error) {
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}

	// Seleccionar archivo de entorno según APP_ENV
	envFile := ".env.dev"
	if appEnv == "production" {
		envFile = ".env"
	}

	// Cargar env específico si está presente (ignorar error si no existe)
	_ = godotenv.Load(envFile)

	// Permitir que el archivo cargado sobreescriba APP_ENV
	if envFromFile := os.Getenv("APP_ENV"); envFromFile != "" {
		appEnv = envFromFile
	}

	cfg := &Config{
		DBSource:  os.Getenv("DB_URL"),
		Port:      os.Getenv("PORT"),
		JWTSecret: os.Getenv("JWT_SECRET"),
		AppEnv:    appEnv,
	}

	if cfg.Port == "" {
		cfg.Port = ":8080"
	}

	if cfg.AppEnv == "" {
		cfg.AppEnv = "development"
	}

	return cfg, nil
}
