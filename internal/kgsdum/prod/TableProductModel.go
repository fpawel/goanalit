package main

import (
	"fmt"
	kgsdum2 "github.com/fpawel/goanalit/internal/kgsdum/kgsdum"
	products2 "github.com/fpawel/goanalit/internal/kgsdum/products"
	"github.com/lxn/walk"
)

type TableProductModel struct {
	walk.ReflectTableModelBase
	product products2.ProductInfo
	db      products2.DB
}

func SetTableProductColumns(cols *walk.TableViewColumnList) {
	check(cols.Clear())

	addColumn("Значение", 80, cols)
	addColumn("Проверка", 200, cols)

	for _, r := range kgsdum2.TestGases() {
		c := walk.NewTableViewColumn()
		check(c.SetTitle(r.String()))
		check(c.SetWidth(80))
		check(cols.Add(c))
	}
}

func (x *TableProductModel) Data(row, col int) (text string,
	image walk.Image, textColor *walk.Color, backgroundColor *walk.Color) {

	test := kgsdum2.Tests()[row/len(kgsdum2.Tests())]
	devVar := kgsdum2.Vars()[row%len(kgsdum2.Vars())]

	switch col {
	case -1:

	case 0:
		text = devVar.String()
	case 1:
		text = test.String()
	default:
		x.db.View(func(tx products2.Tx) {
			if product, f := tx.GetProductByProductTime(x.product.Party.PartyTime, x.product.ProductTime); f {
				value := kgsdum2.Value{
					Product: product,
					Index:   byte(col - 2),
					Test:    test,
					Var:     devVar,
				}.Value()
				if value != nil {
					text = fmt.Sprintf("%g", *value)
				}
			}
		})
	}
	return
}

func addColumn(title string, w int, l *walk.TableViewColumnList) *walk.TableViewColumn {
	c := walk.NewTableViewColumn()
	check(c.SetTitle(title))
	check(c.SetWidth(w))
	check(l.Add(c))
	return c
}

func (x *TableProductModel) StyleCell(c *walk.CellStyle) {
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

func (x *TableProductModel) Value(row, col int) interface{} {
	str, _, _, _ := x.Data(row, col)
	return str
}

func (x *TableProductModel) RowCount() int {
	return len(kgsdum2.Tests()) * len(kgsdum2.Vars())
}
