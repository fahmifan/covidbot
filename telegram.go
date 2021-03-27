package covidbot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
)

type TelegramBotCfg struct {
	Token  string
	Debug  bool
	botAPI *tgbotapi.BotAPI
}

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

func (t *TelegramBot) NotifyAll(text string) error {
	updates, err := t.botAPI.GetUpdates(tgbotapi.UpdateConfig{
		Offset:  0,
		Limit:   100,
		Timeout: 3600,
	})
	if err != nil {
		logrus.Error(err)
		return err
	}

	for _, update := range updates {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
		_, err := t.botAPI.Send(msg)
		if err != nil {
			return err
		}
	}

	return nil
}
