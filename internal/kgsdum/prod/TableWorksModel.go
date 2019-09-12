package main

import (
	kgsdum2 "github.com/fpawel/goanalit/internal/kgsdum/kgsdum"
	products2 "github.com/fpawel/goanalit/internal/kgsdum/products"
	"github.com/lxn/walk"
	"github.com/lxn/win"

	"fmt"
	"log"
)

type TableMainWorksModel struct {
	walk.ReflectTableModelBase
	currentWorkIndex int
	app              *App
}

func NewTableWorksViewModel(app *App) *TableMainWorksModel {
	return &TableMainWorksModel{
		app:              app,
		currentWorkIndex: -1,
	}
}

func (x *TableMainWorksModel) SetCurrentWork(w kgsdum2.Worker) {
	if w == nil {
		x.currentWorkIndex = -1
	} else {
		for i, work := range kgsdum2.Works() {
			if w == work {
				x.currentWorkIndex = i
				break
			}
		}
	}
	x.PublishRowsReset()
}

func (x *TableMainWorksModel) RowCount() int {
	return len(kgsdum2.Works())
}

func (x *TableMainWorksModel) Value(row, col int) (result interface{}) {

	x.app.db.View(func(tx products2.Tx) {
		work := kgsdum2.Works()[row]

		switch col {
		case 0:
			result = work.String()
			return
		case 1:
			r := tx.MostImportantLogRecord(WorkerLogPath(tx, work).Path())
			if r == nil {
				result = ""
				return
			}
			if pt, ok := r.ProductTime(); ok {
				if p, ok := tx.GetCurrentPartyProductByProductTime(pt); ok {
					result = fmt.Sprintf("прибор №%d адр. %d - %s", p.Info().Row, p.Addr(), r.Text)
					return
				}
			}
			result = r.Text
		default:
			log.Panicln("col out of range", col, row)
			result = ""
		}
	})
	return
}

func (x *TableMainWorksModel) Checked(row int) bool {
	_, f := x.app.config.UncheckedWorks[kgsdum2.Works()[row].String()]
	return !f
}

func (x *TableMainWorksModel) SetChecked(row int, checked bool) error {
	k := kgsdum2.Works()[row].String()
	if checked {
		delete(x.app.config.UncheckedWorks, k)
	} else {
		x.app.config.UncheckedWorks[k] = struct{}{}
	}
	return nil
}

func (x *TableMainWorksModel) StyleCell(c *walk.CellStyle) {

	x.app.db.View(func(tx products2.Tx) {
		switch c.Col() {
		case 0:
			if x.currentWorkIndex == c.Row() {
				c.Image = AssetImage("assets/png16/forward.png")
				return
			}
			work := kgsdum2.Works()[c.Row()]
			if r := tx.MostImportantLogRecord(WorkerLogPath(tx, work).Path()); r != nil {
				switch r.Level {
				case win.NIIF_ERROR:
					c.TextColor = walk.RGB(255, 0, 0)
					c.Image = ImgErrorPng16
				case win.NIIF_INFO:
					c.TextColor = walk.RGB(0, 32, 128)
					c.Image = ImgCheckmarkPng16
				}
			}
		}
	})
}

func (x *TableMainWorksModel) ResetCurrentWorkIndex(works []kgsdum2.Worker) {
	x.currentWorkIndex = -1
	x.PublishRowsReset()
}

func WorkerLogPath(tx products2.Tx, x kgsdum2.Worker) products2.DBPath {
	return tx.Party().Test(x.String())
}

func WriteWorkLog(tx products2.Tx, x kgsdum2.Worker, level int, text string) {
	tx.WriteLog(WorkerLogPath(tx, x).Path(), nil, level, text)
}
