package main

import (
	"fmt"
	coef2 "github.com/fpawel/goanalit/internal/coef"
	"github.com/lxn/walk"
	"strconv"
)

type TableCoefsModel struct {
	walk.ReflectTableModelBase
	values coef2.AddrCoefficientValues
}

func NewTableCoefsModel(values coef2.AddrCoefficientValues) (x *TableCoefsModel) {
	return &TableCoefsModel{
		values: values,
	}
}

func (x *TableCoefsModel) RowCount() (n int) {
	return len(x.values.Coefficients())
}
func (x *TableCoefsModel) Value(row, col int) interface{} {
	products := x.values.Addresses()
	productsCount := len(products)
	coefs := x.values.Coefficients()
	coefsCount := len(coefs)
	if col < 0 || row < 0 || col >= productsCount+1 || row >= coefsCount {
		return nil
	}
	k := coefs[row]

	if col == 0 {
		return fmt.Sprintf("%d", k)
	}

	p := products[col-1]

	v, ok := x.values[coef2.AddrCoefficient{Addr: p, Coefficient: k}]
	if !ok {
		return ""
	}
	return strconv.FormatFloat(float64(v), 'f', -1, 32)
}
