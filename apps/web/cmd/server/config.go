package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port    string
	NATSURL string
}

func LoadConfig() *Config {
	// Load root .env file (ignore error if not found)
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		Port:    getEnv("PORT", "8080"),
		NATSURL: getEnv("NATS_URL", "nats://localhost:4222"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
