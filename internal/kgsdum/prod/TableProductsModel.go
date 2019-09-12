package main

import (
	kgsdum2 "github.com/fpawel/goanalit/internal/kgsdum/kgsdum"
	products2 "github.com/fpawel/goanalit/internal/kgsdum/products"
	"github.com/lxn/walk"
	"strconv"

	"errors"
	"fmt"
	"github.com/fpawel/gutils/walkUtils"
	"github.com/lxn/win"
)

type TableProductsModel struct {
	walk.ReflectTableModelBase
	app                  *App
	products             map[products2.ProductTime]*ProductInfo
	surveyRow, surveyCol int
}

func NewTableProductsModel(app *App) (x *TableProductsModel) {
	return &TableProductsModel{
		app:       app,
		surveyRow: -1,
		surveyCol: -1,
		products:  make(map[products2.ProductTime]*ProductInfo),
	}
}

func (x *TableProductsModel) Clear() {
	x.products = make(map[products2.ProductTime]*ProductInfo)
	x.PublishRowsReset()
}

func (x *TableProductsModel) SetSurveyColRow(col, row int) {
	x.surveyRow = row
	x.surveyCol = col
	x.PublishRowsReset()
}

func (x *TableProductsModel) getProductInfo(p products2.ProductTime) *ProductInfo {
	c, ok := x.products[p]
	if !ok {
		c = &ProductInfo{
			values: make(map[kgsdum2.Var]Float32Result),
		}
		x.products[p] = c
	}
	return c
}

func (x *TableProductsModel) SetProductValue(p products2.ProductTime, deviceVar kgsdum2.Var, result Float32Result) {
	c := x.getProductInfo(p)
	m := walkUtils.MessageFromError(result.Error, "успешно")
	c.connection = &m
	c.values[deviceVar] = result
}

func (x *TableProductsModel) SetProductConnection(p products2.ProductTime, err error) {
	c := x.getProductInfo(p)
	m := walkUtils.MessageFromError(err, "ok")
	c.connection = &m
}

func (x *TableProductsModel) AddProductError(p products2.ProductTime, err error) {
	ci := x.getProductInfo(p).connection
	if ci != nil && ci.Level == win.NIIF_ERROR {
		s := ""
		if err != nil {
			s = err.Error() + ", "
		}
		err = errors.New(s + ci.Text)
	}
	x.SetProductConnection(p, err)
}

func (x *TableProductsModel) ClearProductConnection(p products2.ProductTime) {
	x.getProductInfo(p).connection = nil
}

func (x *TableProductsModel) Data(row, col int) (text string, image walk.Image, textColor *walk.Color, backgroundColor *walk.Color) {

	x.app.db.View(func(tx products2.Tx) {
		if row >= len(tx.Party().Products()) {
			return
		}
		p := tx.Party().Products()[row]
		pi := x.getProductInfo(p.ProductTime)

		if col == 0 {

		}
		switch col {
		case 0:
			text = fmt.Sprintf("%02d: %d", p.Addr(), p.Serial())
			if pi.connection == nil {
				return
			}
			switch pi.connection.Level {
			case win.NIIF_ERROR:
				text += ": " + pi.connection.String()
				textColor = new(walk.Color)
				*textColor = walk.RGB(255, 0, 0)
				image = ImgErrorPng16
			case win.NIIF_INFO:
				textColor = new(walk.Color)
				*textColor = walk.RGB(0, 32, 128)
				image = ImgCheckmarkPng16
			}

		default:
			n := col - 1
			if pi, ok := x.products[p.ProductTime]; ok {
				if n >= 0 && n < len(x.app.config.Vars()) {
					if v, ok := pi.values[x.app.config.Vars()[n]]; ok {

						if v.Error == nil {
							text = strconv.FormatFloat(v.Value, 'f', -1, 32)
						} else {
							text = v.Error.Error()
							textColor = new(walk.Color)
							*textColor = walk.RGB(255, 0, 0)
							image = ImgErrorPng16
						}
					}
				}
			}
			if row == x.surveyRow && col == x.surveyCol {
				//image = AssetImage("assets/png16/forward.png")
				backgroundColor = new(walk.Color)
				*backgroundColor = walk.RGB(204, 255, 255)
			}
		}
	})
	return
}

func (x *TableProductsModel) RowCount() (n int) {
	x.app.db.View(func(tx products2.Tx) {
		n = len(tx.Party().Products())
	})
	return
}

func (x *TableProductsModel) StyleCell(c *walk.CellStyle) {
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

func (x *TableProductsModel) Value(row, col int) interface{} {
	str, _, _, _ := x.Data(row, col)
	return str
}

func (x *TableProductsModel) Checked(row int) bool {
	_, ok := x.app.config.UncheckedProducts[byte(row)]
	return !ok
}

func (x *TableProductsModel) SetChecked(row int, checked bool) error {

	k := byte(row)
	if checked {
		delete(x.app.config.UncheckedProducts, k)
	} else {
		x.app.config.UncheckedProducts[k] = struct{}{}
	}
	return nil
}
