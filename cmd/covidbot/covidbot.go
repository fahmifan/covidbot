package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/fahmifan/covidbot"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var timeLayout = "2006-01-02"

func init() {
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
			telegramBot := covidbot.MustTelegramBot(&covidbot.TelegramBotCfg{
				Token: os.Getenv("TELEGRAM_BOT_TOKEN"),
			})
			crawler := &covidbot.PikobarBot{
				APIKey:   os.Getenv("PIKOBAR_API_KEY"),
				Notifier: telegramBot,
			}

			quit := make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt)

			go func() {
				logrus.Info("run crawler everyday..")
				if err := crawler.ScheduleDaily(); err != nil {
					logrus.Fatal(err)
				}
			}()
			<-quit
			logrus.Info("stopped")
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
			fmt.Println(kk.Today())
		},
	}
}
