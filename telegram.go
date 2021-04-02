package covidbot

import (
	"context"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jasonlvhit/gocron"
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
		err := t.Notify(user.ChatID, text)
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

// Notify notify a user
func (t *TelegramBot) Notify(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := t.botAPI.Send(msg)
	if err != nil {
		logrus.WithField("user", Dumps(chatID)).Error(err)
	}
	return err
}

func (t *TelegramBot) NewGoCronSyncUpdates() error {
	err := gocron.Every(1).Hour().Do(t.SyncUpdates)
	if err != nil {
		logrus.Error(err)
	}
	return err
}

// SyncUpdates sync update with telegram bot api
func (t *TelegramBot) SyncUpdates() error {
	logrus.Info("start sync updates")
	offset := 0
	limit := 100
	for {
		updates, err := t.botAPI.GetUpdates(tgbotapi.UpdateConfig{
			Offset: offset,
			Limit:  limit,
		})
		if err != nil {
			logrus.Error(err)
			return err
		}

		if len(updates) == 0 {
			break
		}

		offset = updates[len(updates)-1].UpdateID + 1
		err = t.syncUpdates(updates)
		if err != nil {
			logrus.Error(err)
			return err
		}
	}

	logrus.Info("finished sync updates")
	return nil
}

func (t *TelegramBot) syncUpdates(updates []tgbotapi.Update) error {
	for _, update := range updates {
		if update.Message == nil || update.Message.Chat == nil {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		username := update.Message.Chat.UserName
		chatID := update.Message.Chat.ID
		if username == "" {
			username = fmt.Sprint(chatID)
		}

		user := &User{
			Username:         username,
			TelegramUpdateID: fmt.Sprint(update.UpdateID),
			ChatID:           chatID,
		}
		err := t.UserService.Create(ctx, user)
		if err != nil {
			logrus.WithField("user", Dumps(user)).Error(err)
			return err
		}
	}

	return nil
}
