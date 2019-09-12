package viewmodel

import (
	"github.com/fpawel/daf.v0/internal/assets"
	"github.com/fpawel/daf.v0/internal/data"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type DafProductEntriesTable struct {
	walk.TableModelBase
	party   *data.Party
	product *data.Product
	items   []*data.ProductEntry
}

func NewDafProductEntriesTable() *DafProductEntriesTable {
	return &DafProductEntriesTable{
		party:   new(data.Party),
		product: new(data.Product),
	}
}

func (m *DafProductEntriesTable) SetProduct(productID int64) {

	if err := data.DBProducts.FindByPrimaryKeyTo(m.product, productID); err != nil {
		panic(err)
	}
	if err := data.DBProducts.FindByPrimaryKeyTo(m.party, m.product.PartyID); err != nil {
		panic(err)
	}
	xs, err := data.DBProducts.SelectAllFrom(
		data.ProductEntryTable,
		"WHERE product_id = ? ORDER BY created_at",
		productID)
	if err != nil {
		panic(err)
	}
	m.items = nil
	for _, x := range xs {
		m.items = append(m.items, x.(*data.ProductEntry))
	}
	m.PublishRowsReset()
}

func (m *DafProductEntriesTable) RowCount() int {
	return len(m.items)
}

func (m *DafProductEntriesTable) Value(row, col int) interface{} {
	x := m.items[row]
	switch ProductEntryColumn(col) {

	case ProdEntryColTime:
		return x.CreatedAt

	case ProdEntryColWorkName:
		return x.WorkName

	case ProdEntryColMessage:
		return x.Message

	}
	return ""
}

func (m *DafProductEntriesTable) StyleCell(style *walk.CellStyle) {

	p := m.items[style.Row()]
	switch ProductEntryColumn(style.Col()) {

	case ProdEntryColTime:
		style.TextColor = walk.RGB(0, 128, 0)

	case ProdEntryColMessage:
		if !p.Ok {
			style.TextColor = walk.RGB(255, 0, 0)
			style.Image = assets.ImgError
		} else {
			style.TextColor = walk.RGB(0, 0, 128)
			style.Image = assets.ImgCheckMark
		}
	}
}

type ProductEntryColumn int

const (
	ProdEntryColTime ProductEntryColumn = iota
	ProdEntryColWorkName
	ProdEntryColMessage
)

var ProductEntryColumns = func() []TableViewColumn {
	x := make([]TableViewColumn, ProdEntryColMessage+1)
	type t = TableViewColumn
	x[ProdEntryColTime] =
		t{Title: "Дата", Width: 150, Format: "02.01.06 15:04"}
	x[ProdEntryColWorkName] =
		t{Title: "Проверка", Width: 120}
	x[ProdEntryColMessage] =
		t{Title: "Содержание"}
	return x
}()
