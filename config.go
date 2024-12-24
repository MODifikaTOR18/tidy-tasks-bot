package main

import (
	"log"
	"os"
)

type Config struct {
	TelegramToken string
	DBInfo        DBInfo
}

type DBInfo struct {
	DBUser     string
	DBPassword string
}

func LoadConfig() Config {
	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	if telegramToken == "" {
		log.Fatal("Environment variable TELEGRAM_TOKEN must be set")
	}
	if dbUser == "" {
		log.Fatal("Environment variable DB_USER must be set")
	}
	if dbPassword == "" {
		log.Fatal("Environment variable DB_PASSWORD must be set")
	}

	return Config{
		TelegramToken: telegramToken,
		DBInfo: DBInfo{
			DBUser:     dbUser,
			DBPassword: dbPassword,
		},
	}
}
