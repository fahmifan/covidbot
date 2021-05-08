package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		logrus.Error("Error loading .env file")
	}
}

func PikobarAPIKey() string {
	return os.Getenv("PIKOBAR_API_KEY")
}

func TelegramBotToken() string {
	return os.Getenv("TELEGRAM_BOT_TOKEN")
}

func Port() string {
	return os.Getenv("PORT")
}
