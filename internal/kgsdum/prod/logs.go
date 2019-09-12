package main

import (
	"fmt"
	"github.com/daviddengcn/go-colortext"
	products2 "github.com/fpawel/goanalit/internal/kgsdum/products"
	"github.com/fpawel/gutils/walkUtils"
	"github.com/lxn/walk"
	"github.com/lxn/win"
)

func (x *App) WriteTestProductLog(productTime products2.ProductTime, addr byte, r products2.TestLogRecord) []byte {
	var row int
	x.db.Update(func(tx products2.Tx) {
		var p products2.Product
		var productFound bool
		if addr > 0 {
			p, productFound = tx.GetCurrentPartyProductByAddr(addr)
		} else {
			p, productFound = tx.GetCurrentPartyProductByProductTime(productTime)
			addr = p.Addr()
		}
		if !productFound {
			panic("product not found")
		}
		row = p.Info().Row
		r.TimeKey = p.WriteLog(r)
	})

	text := fmt.Sprintf("Прибор №%d, адрес %d, %s: %s", row, addr, r.Test, r.Text)
	PrintLog(r.Level, text)

	x.mw.Synchronize(func() {
		x.tableLogsModel.PublishRowsReset()
		walkUtils.ScrollDownTableView(x.mw.tblLogs)
		x.mw.lblWorkMessage.SetText(text)
		switch r.Level {
		case win.NIIF_ERROR:
			x.mw.lblWorkMessage.SetTextColor(walk.RGB(255, 0, 0))
		default:
			x.mw.lblWorkMessage.SetTextColor(walk.RGB(0, 102, 204))
		}
	})

	return r.TimeKey
}

func (x *App) WriteLogParty(r products2.TestLogRecord) []byte {
	x.db.Update(func(tx products2.Tx) {
		r.TimeKey = tx.Party().WriteLog(r)
	})

	x.mw.Synchronize(func() {
		x.tableLogsModel.PublishRowsReset()
		walkUtils.ScrollDownTableView(x.mw.tblLogs)
		switch r.Level {
		case win.NIIF_ERROR:
			x.mw.lblWorkMessage.SetTextColor(walk.RGB(255, 0, 0))
		default:
			x.mw.lblWorkMessage.SetTextColor(walk.RGB(0, 102, 204))
		}
		x.mw.lblWorkMessage.SetText(r.Test + ": " + r.Text)
	})
	PrintLog(r.Level, r.Test+": "+r.Text)
	return r.TimeKey
}

func PrintLog(level int, text string) {
	switch level {
	case win.NIIF_ERROR:
		ct.Foreground(ct.Red, true)
		fmt.Println("ERROR:", text)
		ct.ResetColor()
	case win.NIIF_INFO:
		ct.Foreground(ct.Blue, true)
		fmt.Println(text)
		ct.ResetColor()
	default:
		fmt.Println(text)
	}
	return
}
