package viewmodel

import (
	"fmt"
	"github.com/fpawel/daf.v0/internal/assets"
	"github.com/fpawel/daf.v0/internal/data"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type DafProductValuesTable struct {
	walk.TableModelBase
	party   *data.Party
	product *data.Product
	items   []*data.ProductValue
}

func NewDafProductValuesTable() *DafProductValuesTable {
	return &DafProductValuesTable{
		party:   new(data.Party),
		product: new(data.Product),
	}
}

func (m *DafProductValuesTable) SetProduct(productID int64) {

	if err := data.DBProducts.FindByPrimaryKeyTo(m.product, productID); err != nil {
		panic(err)
	}
	if err := data.DBProducts.FindByPrimaryKeyTo(m.party, m.product.PartyID); err != nil {
		panic(err)
	}
	xs, err := data.DBProducts.SelectAllFrom(
		data.ProductValueTable,
		"WHERE product_id = ? ORDER BY work_index",
		productID)
	if err != nil {
		panic(err)
	}
	m.items = nil
	for _, x := range xs {
		m.items = append(m.items, x.(*data.ProductValue))
	}
	m.PublishRowsReset()
}

func (m *DafProductValuesTable) RowCount() int {
	return len(m.items)
}

func (m *DafProductValuesTable) Value(row, col int) interface{} {
	x := m.items[row]
	switch ProductValueColumn(col) {
	case ProdValColWorkIndex:
		return x.WorkIndex + 1
	case ProdValColTime:
		return x.CreatedAt
	case ProdValColGas:
		return fmt.Sprintf("ПГС%d", x.Gas)
	case ProdValColConcentration:
		return x.Concentration
	case ProdValColCurr:
		return x.Current
	}
	return ""
}

func (m *DafProductValuesTable) StyleCell(style *walk.CellStyle) {

	p := m.items[style.Row()]
	switch ProductValueColumn(style.Col()) {

	case ProdValColTime:
		style.TextColor = walk.RGB(0, 128, 0)

	case ProdValColThreshold1:
		if p.Threshold1 {
			style.Image = assets.ImgPinOn
		} else {
			style.Image = assets.ImgPinOff
		}

	case ProdValColThreshold2:
		if p.Threshold2 {
			style.Image = assets.ImgPinOn
		} else {
			style.Image = assets.ImgPinOff
		}
	}
}

type ProductValueColumn int

const (
	ProdValColWorkIndex ProductValueColumn = iota
	ProdValColTime
	ProdValColGas
	ProdValColConcentration
	ProdValColCurr
	ProdValColThreshold1
	ProdValColThreshold2
)

var ProductValueColumns = func() []TableViewColumn {
	x := make([]TableViewColumn, ProdValColThreshold2+1)

	type t = TableViewColumn
	x[ProdValColWorkIndex] =
		t{Title: "№", Width: 50}
	x[ProdValColTime] =
		t{Title: "Дата", Width: 150, Format: "02.01.06 15:04:05"}
	x[ProdValColGas] =
		t{Title: "Газ", Width: 80, Precision: 0}
	x[ProdValColConcentration] =
		t{Title: "Концентрация", Width: 150, Precision: 2}
	x[ProdValColCurr] =
		t{Title: "Ток", Width: 100, Precision: 1}
	x[ProdValColThreshold1] =
		t{Title: "Порог 1", Width: 120}
	x[ProdValColThreshold2] =
		t{Title: "Порог 2", Width: 120}
	return x
}()
