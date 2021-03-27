package covidbot

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/jasonlvhit/gocron"
	"github.com/sirupsen/logrus"
)

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

// ScheduleDaily run CarwlDailiy every day and send notification
func (p *PikobarBot) ScheduleDaily() error {
	dir := "crawl-output"
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	timeLayout := "2006-01-02"
	err = gocron.Every(1).Day().At("11:05").Do(func() error {
		fileName := time.Now().Format(timeLayout) + ".json"
		dst := path.Join(dir, fileName)
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

		logrus.Info("send notifications")
		err = p.Notifier.NotifyAll(kk.Today().String())
		if err != nil {
			logrus.Error(err)
			return err
		}

		logrus.Info("success notify all users")
		return nil
	})
	if err != nil {
		return err
	}

	gocron.RunAll()
	return nil
}
