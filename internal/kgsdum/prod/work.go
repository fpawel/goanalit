package main

import (
	"fmt"
	products2 "github.com/fpawel/goanalit/internal/kgsdum/products"
	"github.com/fpawel/guartutils/comport"
	"github.com/fpawel/gutils/walkUtils"
	"github.com/lxn/walk"
	"sync/atomic"
	"time"
)

type ProductDataReader func(p products2.ProductInfo) error

func (x *App) PortProducts() comport.Fetcher {
	return x.ports.Port(x.config.SerialPorts.PortProducts)
}

func (x *App) OpenPorts(ports []ComportType) error {
	x.ports.Close()
	for _, p := range ports {
		c := x.config.SerialPorts.Port(p).Port
		err := x.ports.Open(c)
		if err != nil {
			return fmt.Errorf("ошибка открытия порта %s: %v", c.Name, err)
		}
	}
	return nil
}

func (x *App) ProcessEachProduct(what, str2 string, productDataReader ProductDataReader) error {
	defer x.mw.tblProducts.Synchronize(func() {
		x.tableProductsModel.SetSurveyColRow(-1, -1)
		x.tableLogsModel.PublishRowsReset()
	})

	var errorProducts ErrorProducts

	for row, p := range x.db.Products() {
		p := p
		if !x.tableProductsModel.Checked(row) {
			continue
		}
		x.mw.tblProducts.Synchronize(func() {
			x.tableProductsModel.SetSurveyColRow(-1, p.Row)
		})
		err := productDataReader(p)

		if err == fetch.ErrorCanceled {
			return fetch.ErrorCanceled
		}

		x.mw.tblProducts.Synchronize(func() {
			x.tableProductsModel.SetProductConnection(p.ProductTime, err)
		})

		text, level := walkUtils.ErrorLevel(err, str2+": успешно")
		x.WriteTestProductLog(p.ProductTime, 0, products2.TestLogRecord{
			// what, nil, level, text
			Test:  what,
			Level: level,
			Text:  text,
		})

		textColor := walk.RGB(0, 102, 204)
		if err != nil {
			textColor = walk.RGB(255, 0, 0)
		}
		x.mw.Synchronize(func() {
			x.mw.lblWorkMessage.SetTextColor(textColor)
			x.mw.lblWorkMessage.SetText(what + ": " + text)
		})
		errorProducts = append(errorProducts, ErrorProduct{p, err})
	}
	if errorProducts.HasErrors() {
		return errorProducts
	}
	return nil
}

func (x *App) Delay(what string, duration time.Duration, work func()) {
	startTime := time.Now()
	x.cancellationDelay = 0
	mw := x.mw

	mw.Synchronize(func() {
		mw.progressBar.SetVisible(true)
		mw.progressBar.SetRange(0, int(duration.Nanoseconds()/1000000))
		mw.progressBar.SetValue(0)
		mw.btnCancelDelay.SetVisible(true)
	})

	for atomic.LoadInt32(&x.cancellationDelay) == 0 && time.Since(startTime) < duration {
		mw.Synchronize(func() {
			mw.progressBar.SetValue(int(time.Since(startTime).Nanoseconds() / 1000000))
			str := fmt.Sprintf("%s, задержка %v - %v", what, time.Since(startTime), duration)
			check(mw.lblWorkMessage.SetText(str))
		})
		work()
	}
	mw.Synchronize(func() {
		mw.progressBar.SetVisible(false)
		mw.btnCancelDelay.SetVisible(false)
	})
}
