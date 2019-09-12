package productsView

import (
	"fmt"
	products2 "github.com/fpawel/goanalit/internal/kgsdum/products"
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	"time"
)

type TableLogsViewModel struct {
	walk.ReflectTableModelBase
	db        products2.DB
	partyTime products2.PartyTime
}

func NewTableLogsViewModel(db products2.DB, partyTime products2.PartyTime) *TableLogsViewModel {
	return &TableLogsViewModel{
		db:        db,
		partyTime: partyTime,
	}
}

func TableLogsColumns() []declarative.TableViewColumn {
	return []declarative.TableViewColumn{
		{Title: "Дата", Width: 100},
		{Title: "Время", Width: 100},
		{Title: "Прибор", Width: 150},
		{Title: "Действие", Width: 400},
		{Title: "Сообщение", Width: 500},
	}
}

func (x *TableLogsViewModel) RowCount() (result int) {
	var logs products2.Logs
	x.db.View(func(tx products2.Tx) {
		result = len(tx.Logs(x.partyTime.Path(), logs))
	})
	return
}
func (x *TableLogsViewModel) Data(row, col int) (text string, image walk.Image, textColor *walk.Color, backgroundColor *walk.Color) {

	return viewData(x.db, nil, x.partyTime, row, col)
}

func (x *TableLogsViewModel) Value(row, col int) (r interface{}) {
	str, _, _, _ := x.Data(row, col)
	return str
}

func (x *TableLogsViewModel) StyleCell(c *walk.CellStyle) {
	_, image, textColor, backgroundColor := x.Data(c.Row(), c.Col())
	if image != nil {
		c.Image = image
	}
	if textColor != nil {
		c.TextColor = *textColor
	}
	if backgroundColor != nil {
		c.BackgroundColor = *backgroundColor
	}
}

func viewData(db products2.DB, logs products2.Logs, partyTime products2.PartyTime, row, col int) (text string,
	image walk.Image, textColor *walk.Color, backgroundColor *walk.Color) {
	db.View(func(tx products2.Tx) {
		var logRec *products2.LogRecord
		var logTm time.Time
		partyPath := partyTime.Path()
		logs := tx.Logs(partyPath, logs)
		for i, t := range logs.Times() {
			if i == row {
				logRec = logs[t]
				logTm = t
				break
			}
		}
		if logRec == nil {
			return
		}
		switch col {
		case 0:
			text = logTm.Format("02.01.2006")
		case 1:
			text = logTm.Format("15:04:05")
		case 2:

			var product products2.Product
			productTime, foundProduct := logRec.ProductTime()

			partyTime := products2.KeyToTime(partyPath[len(partyPath)-1])
			product = products2.Product{
				ProductTime: productTime,
				Party: products2.Party{
					Tx:        tx,
					PartyTime: products2.PartyTime(partyTime),
				},
			}
			if foundProduct {
				text = fmt.Sprintf("адрес %02d, № %d", product.Addr(), product.Info().Row+1)
			}

		case 3:
			for i := 0; i < len(logRec.Path)-1; i++ {
				if string(logRec.Path[i]) == "tests" {
					text = string(logRec.Path[i+1])
					return
				}
			}
		case 4:
			text = logRec.Text
			switch logRec.Level {
			case win.NIIF_INFO:
				image = ImgCheckmarkPng16
			case win.NIIF_ERROR:
				image = ImgErrorPng16
				textColor = new(walk.Color)
				*textColor = walk.RGB(255, 0, 0)
			}
		}
	})
	return
}
