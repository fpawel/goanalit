package internal

import (
	"github.com/lxn/walk"
)

const (
	PartyProductsCount   = 50
	StendNotConnceted    = 0
	StendConnceted       = 1
	StendConnectionError = 2
)

type PartyModel struct {
	walk.ReflectTableModelBase
	products [PartyProductsCount]Product
	conn     [5]int
	errCount [5]int
	chart    *Chart
	serials  [PartyProductsCount]uint
}

type ConnStend struct {
}

type Product struct {
	value *float32
	err   bool
}

// Called by the TableView from SetModel and every time the modbusViewModel publishes a
// RowsReset event.
func (x *PartyModel) RowCount() int {
	return PartyProductsCount / 5
}

// Called by the TableView when it needs the text to display for a given cell.
func (x *PartyModel) Value(row, col int) interface{} {

	i := (col/3)*10 + row
	if i < 0 || i >= PartyProductsCount {
		logger.Panicf("unexpected col %d row %d", col, row)
	}

	p := x.products[i]
	switch col % 3 {
	case 0:
		return i + 1

	case 1:
		return x.serials[i]

	default:

		if x.conn[i/10] != StendConnceted {
			return ""
		}

		if p.err {
			return "?"
		}

		if p.value == nil {
			return ""
		}

		if *p.value == 112 {
			return "-"
		}

		return *p.value
	}
}
