package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/fahmifan/covidbot"
	"github.com/fahmifan/covidbot/bbolt"
	"github.com/fahmifan/covidbot/http"
	"github.com/jasonlvhit/gocron"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var timeLayout = "2006-01-02"

func init() {
	logrus.SetReportCaller(true)
	err := godotenv.Load()
	if err != nil {
		logrus.Error("Error loading .env file")
	}
}

var rootCMD = &cobra.Command{
	Short: "covidbotd command line tool",
}

func crawlerCMD() *cobra.Command {
	cmdCrawler := &cobra.Command{
		Use:   "crawl",
		Short: "crawl into pikobar api and output a json",
		Run: func(cmd *cobra.Command, args []string) {
			crawler := covidbot.PikobarBot{
				APIKey: os.Getenv("PIKOBAR_API_KEY"),
			}
			filename := time.Now().Format(timeLayout)
			f, err := os.Create(fmt.Sprintf("%s.json", filename))
			if err != nil {
				logrus.Fatal(err)
			}
			defer f.Close()

			err = crawler.CrawlDailyCase(f)
			if err != nil {
				logrus.Fatal(err)
			}

			logrus.Info("success crawl")
		},
	}

	cmdEveryday := &cobra.Command{
		Use:   "everyday",
		Short: "craw every day",
		Run: func(cmd *cobra.Command, args []string) {
			boltDB := covidbot.MustOpenBoltDB()
			userService := bbolt.NewUserService(&bbolt.UserServiceCfg{
				DB: boltDB,
			})
			telegramBot := covidbot.MustTelegramBot(&covidbot.TelegramBotCfg{
				Token:       os.Getenv("TELEGRAM_BOT_TOKEN"),
				UserService: userService,
			})
			crawler := &covidbot.PikobarBot{
				APIKey:   os.Getenv("PIKOBAR_API_KEY"),
				Notifier: telegramBot,
			}
			server := http.NewServer(&http.Config{
				Port:        os.Getenv("PORT"),
				UserService: userService,
			})

			if err := crawler.NewGoCronDaily(); err != nil {
				logrus.Error(err)
			}
			if err := telegramBot.NewGoCronSyncUpdates(); err != nil {
				logrus.Error(err)
			}

			// run services
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt)
			go func() {
				if err := server.Run(); err != nil {
					logrus.Error(err)
				}
			}()
			var stopCron chan bool
			go func() {
				logrus.Info("run cron job")
				stopCron = gocron.Start()
			}()

			// block main goroutine
			<-quit
			stopCron <- true

			ctx, finish := context.WithTimeout(context.Background(), time.Second*30)
			defer finish()
			if err := server.Stop(ctx); err != nil {
				logrus.Error(err)
			}
		},
	}

	cmdCrawler.AddCommand(cmdEveryday)
	return cmdCrawler
}

func parseCMD() *cobra.Command {
	return &cobra.Command{
		Use:   "parse",
		Short: "parse json output from crawl",
		Run: func(cmd *cobra.Command, args []string) {
			filename := time.Now().Format(timeLayout)
			f, err := os.Open(fmt.Sprintf("%s.json", filename))
			if err != nil {
				logrus.Fatal(err)
			}
			defer f.Close()

			cc, err := covidbot.NewCovidCase(f)
			if err != nil {
				logrus.Fatal(err)
			}

			bandungKode := "3273"
			kk := cc.FilterKabKots(bandungKode)
			now := time.Now()
			oneDay := time.Hour * 24
			fmt.Print(
				kk.Date(now.Format(timeLayout)),
				kk.Date(now.Add(-oneDay).Format(timeLayout)),
				kk.Date(now.Add(-2*oneDay).Format(timeLayout)),
			)
		},
	}
}

func testerCMD() *cobra.Command {
	cmdTester := &cobra.Command{Use: "test"}

	cmdNotify := &cobra.Command{Use: "notify", Short: "notify telegram user"}
	cmdNotify.Flags().Int64("chatID", 0, "--chatID 123")
	cmdNotify.Run = func(cmd *cobra.Command, args []string) {
		chatID := covidbot.StringToInt64(cmd.Flag("chatID").Value.String())
		if chatID <= 0 {
			logrus.WithField("chatID", chatID).Error("invalid chatID")
			return
		}

		telegramBot := covidbot.MustTelegramBot(&covidbot.TelegramBotCfg{
			Token: os.Getenv("TELEGRAM_BOT_TOKEN"),
		})

		err := telegramBot.Notify(chatID, "test")
		if err != nil {
			logrus.WithField("chatID", chatID).Error(err)
			return
		}
		logrus.Info("success")
	}

	cmdTester.AddCommand(cmdNotify)
	return cmdTester
}
