package internal

import (
	"github.com/boltdb/bolt"

	"errors"
	"time"
)

type ModelConfig struct {
	appFolderPath string
	mainWindow    *AppMainWindow
}

type Model struct {
	db     *bolt.DB
	party  PartyModel
	config *ModelConfig
}

var ErrorPartyIDnotFound error = errors.New("PartyModel ID not found")

func NewModel(config *ModelConfig) (x *Model) {

	db, err := bolt.Open(config.appFolderPath+"/chart.db", 0600, nil)
	if err != nil {
		logger.Panic(err)
	}
	x = &Model{
		db:     db,
		config: config,
	}
	err = db.Update(func(tx *bolt.Tx) error {

		buckCharts, err := tx.CreateBucketIfNotExists([]byte("charts"))
		if err != nil {
			logger.Panic(err)
		}

		var lastPartyUpdatedAt time.Time

		// запомнить партии
		c := buckCharts.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if v == nil {
				chart := &Chart{CreatedAt: time.Unix(0, btoi(k))}
				updatedAt, err := chart.getUpdatedAt(tx)

				if updatedAt == chart.CreatedAt {
					buckCharts.DeleteBucket(k)
				} else {
					if err != nil {
						logger.Panic(err)
					}
					if updatedAt.UnixNano() > lastPartyUpdatedAt.UnixNano() && time.Since(updatedAt) < 2*time.Minute {
						x.party.chart = chart
					}
				}
			}
		}
		if x.party.chart == nil {
			x.party.chart = &Chart{CreatedAt: time.Now()}
		}
		return nil
	})
	if err != nil {
		logger.Panic(err)
	}
	return
}

func (x *Model) GetSavedCharts() (charts []*Chart) {
	err := x.db.View(func(tx *bolt.Tx) error {

		buckCharts := tx.Bucket([]byte("charts"))
		if buckCharts == nil {
			return nil
		}

		c := buckCharts.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if v == nil {
				chart := &Chart{CreatedAt: time.Unix(0, btoi(k))}
				charts = append(charts, chart)
			}
		}
		return nil
	})
	if err != nil {
		logger.Panic(err)
	}
	return
}

func (x *Model) OpenChart(chart *Chart) {
	err := x.db.View(func(tx *bolt.Tx) error {
		x.RestorePartySeries(chart, x.config.mainWindow.chart)
		return nil
	})
	if err != nil {
		logger.Panic(err)
	}
	return
}

func (x *Model) StoreTimeValues(xs []TimeValue) {
	if len(xs) == 0 {
		return
	}
	err := x.db.Update(func(tx *bolt.Tx) error {
		for _, timeValue := range xs {
			if err := x.party.chart.storeTimeValue(tx, timeValue); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		logger.Panic(err)
	}
}

func (x *Model) GetPartyUpdatedAt(p *Chart) (t time.Time) {
	err := x.db.View(func(tx *bolt.Tx) (err error) {
		t, err = p.getUpdatedAt(tx)
		return
	})
	if err != nil {
		logger.Panic(err)
	}
	return
}

// загрузить сохранённые точки в чарт
func (x *Model) RestorePartySeries(party *Chart, chart *PipeChartServer) {
	var restoredTimeValues []TimeValue
	err := x.db.View(func(tx *bolt.Tx) (err error) {
		restoredTimeValues, err = party.restoreSeries(tx)
		return
	})
	if err != nil {
		logger.Panic(err)
	}

	mw := x.config.mainWindow

	mw.panelOpenChart.Synchronize(func() {
		mw.panelOpenChart.SetVisible(true)
		mw.progressOpenChart.SetRange(0, len(restoredTimeValues))
		mw.progressOpenChart.SetValue(0)
		mw.panelOpenChart.Invalidate()
	})

	for i := range restoredTimeValues {
		timeValue := restoredTimeValues[i]
		timeValue.N += 52
		chart.SendTimeValue(PipeChartCmdRestoreTimeValue, timeValue)
		if i%1000 == 0 {
			mw.Synchronize(func() {
				mw.progressOpenChart.SetValue(i)
			})
		}
	}
	chart.Send(PipeChartCmdSetRadiogroup1Itemindex, []byte{1})

	mw.panelOpenChart.Synchronize(func() {
		mw.panelOpenChart.SetVisible(false)
	})

	return
}
