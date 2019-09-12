package main

import (
	"fmt"
	products2 "github.com/fpawel/goanalit/internal/kgsdum/products"
	"github.com/lxn/walk"
	"time"
)

type AppMainWindow struct {
	*walk.MainWindow
	app            *App
	tblProducts    *walk.TableView
	tblWorks       *walk.TableView
	tblLogs        *walk.TableView
	btnRun         *walk.SplitButton
	btnCancel      *walk.PushButton
	btnCancelDelay *walk.PushButton
	lblWorkInfo    *walk.Label
	lblWorkMessage *walk.TextEdit
	progressBar    *walk.ProgressBar
}

func NewAppMainWindow(app *App) *AppMainWindow {
	return &AppMainWindow{
		app: app,
	}
}

func (x *AppMainWindow) Initialize() {
	x.progressBar.SetVisible(false)
	x.btnCancel.SetVisible(false)
	x.btnCancelDelay.SetVisible(false)
	x.app.db.View(func(tx products2.Tx) {
		t := time.Time(tx.Party().PartyTime)
		check(x.SetTitle(fmt.Sprintf("%s Партия %s",
			x.app.ProductName(), t.Format("02 01 2006, 03:04"))))
	})
	x.SetupProductsColumns()
}

func (x *AppMainWindow) InvalidateWorkRunning(isRunning bool) {
	x.btnRun.SetVisible(!isRunning)
	x.btnCancel.SetVisible(isRunning)
	acts := x.btnRun.Menu().Actions()
	for i := 0; i < acts.Len(); i++ {
		check(acts.At(i).SetVisible(!isRunning))
	}
	x.app.tableProductsModel.SetSurveyColRow(-1, -1)
}

func (x *AppMainWindow) SetupProductsColumns() {
	cols := x.tblProducts.Columns()
	for cols.Len() > 1 {
		check(cols.RemoveAt(cols.Len() - 1))
	}

	for _, v := range x.app.config.Vars() {
		col := walk.NewTableViewColumn()
		check(col.SetTitle(v.String()))
		check(col.SetWidth(80))
		check(cols.Add(col))
	}
}
