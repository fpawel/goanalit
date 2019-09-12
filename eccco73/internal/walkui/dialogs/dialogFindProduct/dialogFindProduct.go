package dialogFindProduct

import (
	"fmt"
	"github.com/fpawel/eccco73/internal/eccco73"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	"log"
	"time"
)

type FinderProductsBySerial interface {
	FindProductsBySerial(serial int64) (xs []eccco73.FoundProduct)
}

func Execute(owner walk.Form, finder FinderProductsBySerial, serial int64) (p eccco73.FoundProduct, ok bool) {

	var dlg *walk.Dialog
	var edSerial *walk.NumberEdit
	var tableView *walk.TableView
	var lblFind *walk.Label

	tableModel := &tableModel{ }
	table := TableView{
		AssignTo:              &tableView,
		LastColumnStretched:   true,
		AlternatingRowBGColor: walk.RGB(239, 239, 239),
		ColumnsOrderable:      false,
		MultiSelection:        false,
		Columns: []TableViewColumn{
			{
				Title: "Дата",
				Width: 90,
			},
			{
				Title: "Исполнение",
				Width: 85,
			},
			{
				Title: "Партия",
				Width: 260,
			},
			{
				Title: "Прошивка",
			},
			{
				Title: "Примечание",
			},
		},
		Model: tableModel,
		OnItemActivated: func() {

			n := tableView.CurrentIndex()
			if n >= 0 && n < len(tableModel.products) {
				p = tableModel.products[n]
				dlg.Accept()
			}
		},
	}

	var timer *time.Timer
	var done chan bool
	doFind := func() {
		for {
			select {
			case <-done:
				return
			case <-timer.C:
				lblFind.Synchronize(func() {
					lblFind.SetVisible(true)
				})
				tableModel.products = finder.FindProductsBySerial(int64(edSerial.Value()))
				dlg.Synchronize(func() {
					lblFind.SetVisible(false)
					tableModel.PublishRowsReset()
				})
			}
		}
	}
	debounceFind := func(){
		if timer != nil {
			timer.Stop()
			done <- true
			lblFind.SetVisible(false)
		}
		done = make(chan bool,2)
		timer = time.NewTimer(500 * time.Millisecond)
		if err :=lblFind.SetText(fmt.Sprintf("Поиск %v...", edSerial.Value())); err != nil {
			panic(err)
		}

		go doFind()
	}

	rightPanel := Composite{
		Layout: VBox{},
		Children: []Widget{

			Label{Text: "Заводской номер:"},

			NumberEdit{
				AssignTo: &edSerial,
				Value:    float64(serial),
				OnValueChanged: debounceFind,
			},

			Label{
				AssignTo:&lblFind,
				Text:"Поиск...",
				Visible:false,
			},
			ScrollView{
				HorizontalFixed: true,
				Layout:          VBox{},
			},
		},
	}

	dlgm := Dialog{
		Icon:          "assets/ico/settings.ico",
		AssignTo:      &dlg,
		Title:         "Поиск ЭХЯ по номеру",
		Font:          Font{PointSize: 10},
		MinSize:       Size{800, 500},
		MaxSize:       Size{800, 500},
		FixedSize:     true,

		Layout: HBox{},
		Children: []Widget{
			table,
			rightPanel,
		},
	}

	if err := dlgm.Create(owner); err != nil {
		log.Panic(err)
	}
	ok = dlg.Run() == win.IDOK
	if done != nil {
		done <- true
	}
	return
}

type tableModel struct {
	walk.ReflectTableModelBase
	products []eccco73.FoundProduct
}

func (x *tableModel) RowCount() int {
	return len(x.products)
}

func (x *tableModel) Value(row, col int) interface{} {
	p := x.products[row]
	switch col {
	case 0:
		return p.CreatedAt.Format("02.01.2006")
	case 1:
		return p.ProductTypeName()
	case 2:
		return p.PartyNote.String
	case 4:
		return p.Note.String

	default:
		return ""
	}
}

func (x *tableModel) StyleCell(c *walk.CellStyle) {
	if c.Col() == 3 {
		p := x.products[c.Row()]
		if p.Flash {
			c.Image = "assets/png16/checkmark.png"
		}
	}
}