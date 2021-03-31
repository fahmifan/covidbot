package covidbot

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
)

// TelegramBotCfg ..
type TelegramBotCfg struct {
	botAPI *tgbotapi.BotAPI

	UserService UserService
	Token       string
	Debug       bool
}

// TelegramBot ..
type TelegramBot struct {
	*TelegramBotCfg
}

// NewTelegramBot ..
func NewTelegramBot(cfg *TelegramBotCfg) (*TelegramBot, error) {
	bot := &TelegramBot{cfg}
	botAPI, err := tgbotapi.NewBotAPI(bot.Token)
	if err != nil {
		return nil, err
	}
	botAPI.Debug = cfg.Debug
	bot.botAPI = botAPI
	return bot, nil
}

// MustTelegramBot init TelegramBot fatal on error
func MustTelegramBot(cfg *TelegramBotCfg) *TelegramBot {
	bot, err := NewTelegramBot(cfg)
	if err != nil {
		logrus.Fatal(err)
	}
	return bot
}

// NotifyAll notify all users
func (t *TelegramBot) NotifyAll(text string) error {
	err := t.UserService.ForEach(context.Background(), func(user User) error {
		msg := tgbotapi.NewMessage(user.ChatID, text)
		_, err := t.botAPI.Send(msg)
		if err != nil {
			logrus.WithField("user", Dumps(user)).Error(err)
		}
		return nil
	})
	if err != nil {
		logrus.Error(err)
	}

	return err
}
