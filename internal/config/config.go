package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Postgres Database
	App      Application
	Log      Logger
}

type Database struct {
	Host     string
	Port     string
	Username string
	Password string
	Name     string
	Mode     string
}

type Application struct {
	Port string
	Env  string
}

type Logger struct {
	Level string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println(".env файл не найден, читаем переменные окружения")
	}

	cfg := &Config{
		Postgres: Database{
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			Username: os.Getenv("DB_USERNAME"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     os.Getenv("DB_NAME"),
			Mode:     os.Getenv("DB_MODE"),
		},
		App: Application{
			Port: getEnvOrDefault("APP_PORT", "8080"),
			Env:  getEnvOrDefault("APP_ENV", "production"),
		},
		Log: Logger{
			Level: getEnvOrDefault("LOG_LEVEL", "info"),
		},
	}

	return cfg
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
