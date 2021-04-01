package covidbot

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

const timeLayout = "2006-01-02"

type Case struct {
	ClosecontactDikarantina int     `json:"closecontact_dikarantina"`
	ClosecontactDiscarded   int     `json:"closecontact_discarded"`
	ClosecontactRatarata    float64 `json:"closecontact_ratarata"`
	ClosecontactTotal       int     `json:"closecontact_total"`
	ConfirmationDiisolasi   int     `json:"confirmation_diisolasi"`
	ConfirmationMeninggal   int     `json:"confirmation_meninggal"`
	ConfirmationRatarata    float64 `json:"confirmation_ratarata"`
	ConfirmationSelesai     int     `json:"confirmation_selesai"`
	ConfirmationTotal       int     `json:"confirmation_total"`
	ProbableDiisolasi       int     `json:"probable_diisolasi"`
	ProbableDiscarded       int     `json:"probable_discarded"`
	ProbableMeninggal       int     `json:"probable_meninggal"`
	ProbableRatarata        float64 `json:"probable_ratarata"`
	ProbableTotal           int     `json:"probable_total"`
	SuspectDiisolasi        int     `json:"suspect_diisolasi"`
	SuspectDiscarded        int     `json:"suspect_discarded"`
	SuspectMeninggal        int     `json:"suspect_meninggal"`
	SuspectRatarata         float64 `json:"suspect_ratarata"`
	SuspectTotal            int     `json:"suspect_total"`
}

type Item struct {
	KodeKab   string `json:"kode_kab"`
	NamaKab   string `json:"nama_kab"`
	Tanggal   string `json:"tanggal"`
	Harian    Case   `json:"harian"`
	Kumulatif Case   `json:"kumulatif"`
}

func (d Item) String() string {
	return fmt.Sprintf("Tgl: %s, Konfirmasi Baru: %d, Kasus Aktif: %d\n",
		d.Tanggal,
		d.Harian.ConfirmationDiisolasi,
		d.Kumulatif.ConfirmationDiisolasi,
	)
}

type KabKot struct {
	KodeKab string `json:"kode_kab"`
	NamaKab string `json:"nama_kab"`
	Series  []Item `json:"series"`
}

// Today find today's item
func (kk KabKot) Today() Item {
	return kk.Date(time.Now().Format(timeLayout))
}

// Date find an Item by date
func (kk KabKot) Date(date string) Item {
	for i := range kk.Series {
		if kk.Series[i].Tanggal == date {
			return kk.Series[i]
		}
	}
	return Item{}
}

type CovidCase struct {
	KabKots []KabKot `json:"data"`
}

func (d CovidCase) FilterKabKots(kode string) KabKot {
	for i := range d.KabKots {
		if d.KabKots[i].KodeKab == kode {
			return d.KabKots[i]
		}
	}

	return KabKot{}
}

// NewCovidCase populate CovidCase from reader
func NewCovidCase(rd io.Reader) (CovidCase, error) {
	dc := CovidCase{}
	err := json.NewDecoder(rd).Decode(&dc)
	if err != nil {
		return dc, fmt.Errorf("failed to decode: %w", err)
	}

	return dc, nil
}
