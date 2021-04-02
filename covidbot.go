package covidbot

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/jasonlvhit/gocron"
	"github.com/sirupsen/logrus"
)

const defaultDir = "crawl-output"

type Notifier interface {
	NotifyAll(text string) error
}

// PikobarBot ..
type PikobarBot struct {
	APIKey   string
	Notifier Notifier
}

// CrawlDailyCase ..
func (p *PikobarBot) CrawlDailyCase(wr io.Writer) error {
	host := "https://dashboard-pikobar-api.digitalservice.id/v2/kasus/harian?wilayah=kota"
	return p.Crawl(host, wr)
}

// Crawl crawl data from the host and write into wr
func (p *PikobarBot) Crawl(host string, wr io.Writer) error {
	req, err := http.NewRequest(http.MethodGet, host, nil)
	if err != nil {
		logrus.Fatal(err)
	}
	header := http.Header{}
	header.Add("api-key", p.APIKey)
	req.Header = header

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		logrus.Fatal(err)
	}
	defer res.Body.Close()

	bt, err := io.ReadAll(res.Body)
	if err != nil {
		logrus.Fatal(err)
	}

	if res.StatusCode != http.StatusOK {
		logrus.WithField("status", res.StatusCode).Error(string(bt))
		return fmt.Errorf("status not ok: %d", res.StatusCode)
	}

	_, err = wr.Write(bt)
	return err
}

// NewGoCronDaily set cron every day and send notification
func (p *PikobarBot) NewGoCronDaily() error {
	err := os.MkdirAll(defaultDir, os.ModePerm)
	if err != nil {
		return err
	}

	return gocron.Every(1).Days().At("09:00").Do(p.crawlAndNotify)
}

func (p *PikobarBot) crawlAndNotify() error {
	fileName := time.Now().Format(timeLayout) + ".json"
	dst := path.Join(defaultDir, fileName)
	file, err := os.Create(dst)
	if err != nil {
		logrus.Error(err)
		return err
	}
	defer file.Close()

	buff := bytes.NewBuffer(nil)
	err = p.CrawlDailyCase(buff)
	if err != nil {
		logrus.Error(err)
		file.Close()
		return err
	}

	_, err = file.Write(buff.Bytes())
	if err != nil {
		logrus.Error(err)
		return err
	}

	logrus.Info("read crawl data")
	cc, err := NewCovidCase(buff)
	if err != nil {
		logrus.Error(err)
		return err
	}

	bandungKode := "3273"
	kk := cc.FilterKabKots(bandungKode)
	now := time.Now()
	oneDay := time.Hour * 24
	last3Days := []Item{
		kk.Date(now.Format(timeLayout)),
		kk.Date(now.Add(-oneDay).Format(timeLayout)),
		kk.Date(now.Add(-2 * oneDay).Format(timeLayout)),
	}
	notifMsg := strings.Join([]string{
		last3Days[0].String(),
		last3Days[1].String(),
		last3Days[2].String(),
	}, "\n")
	notifMsg = fmt.Sprintf("Update from Last 3 days\n\n%s", notifMsg)

	logrus.WithField("msgs", notifMsg).Info("send notifications")

	err = p.Notifier.NotifyAll(notifMsg)
	if err != nil {
		logrus.Error(err)
		return err
	}

	logrus.Info("success notify all users")
	return nil
}

func StringToInt64(s string) int64 {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return v
}
