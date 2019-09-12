package main

import (
	"database/sql"
	"fmt"
	"github.com/fpawel/eccco73/internal/eccco73"
	"github.com/lxn/walk"
	"strconv"
)

type ProductsTableModel struct {
	walk.ReflectTableModelBase
	app *App
}

type productColumn int

const (
	colNum productColumn = iota
	colProdType
	colSerial
	colFon20
	colSens20
	colCalcKSens20

	colFonMinus20
	colSensMinus20

	colFon50
	colSens50
	colCalcKSens50

	colCalcDFon
	colCalcDT

	colGas13
	colGas24
	colGas35
	colGas26
	colGas17

	colNei
	colCalcDNei

	colNote
)

var productColumns = [colNote + 1]string{
	"№", "ИБЯЛ", "Зав.№",
	"фон.20", "Ч.20", "Кч.20",
	"фон.-20", "Ч.-20",
	"фон.50", "Ч.50", "Kч.50",

	"D.фон", "D.T",

	"ПГС1", "ПГС2", "ПГС3", "ПГС2", "ПГС1",

	"неизм.", "D.неи",

	"Примечание",
}

func (x *ProductsTableModel) RowCount() int {
	return len(x.app.products)
}

func (x *ProductsTableModel) Data(row, col int) (text string, image string, textColor, backgroundColor walk.Color,
	applyTextColor, applyBackgroundColor bool) {
	p := x.app.products[row]
	p2 := eccco73.Product2{
		Product: p,
		Party:   x.app.party,
		Types:   x.app.productTypes,
	}

	if x.app.mw.clickedProduct.ProductID == p.ProductID {
		textColor = walk.RGB(0, 0, 200)
		backgroundColor = walk.RGB(208, 211, 212)
		applyTextColor = true
		applyBackgroundColor = true
	}

	setError := func() {
		image = "assets/png16/warning.png"
		textColor = walk.RGB(255, 0, 0)
		applyTextColor = true
	}

	switch productColumn(col) {
	case colNum:
		text = fmt.Sprintf("%d.%d", p.Order/8+1, p.Order%8+1)
		if x.app.mw.clickedProduct.ProductID == p.ProductID {
			image = "assets/png16/forward.png"
			textColor = walk.RGB(0, 0, 255)
			applyTextColor = true
		} else {
			if len(p.FlashBytes) > 0 {
				image = Png16Checkmark
			}
		}

	case colProdType:
		if p.ProductTypeID.Valid {
			text = p2.ProductType().Name
		}
		if !p2.Ok() {
			setError()
		}

	case colSerial:
		text = fmt.Sprintf("%d", p.Serial)

	case colNote:
		if p.Note.Valid {
			text = p.Note.String
		}
	case colFon20:
		text = fmtNullFloat64(p.Fon20, 2)
		if !p2.OkFon20() {
			setError()
		}
	case colSens20:
		text = fmtNullFloat64(p.Sens20, 2)
	case colGas13:
		text = fmtNullFloat64(p.I13, 2)
		if !p2.OkDeltaFon20() {
			setError()
		}
	case colFonMinus20:
		text = fmtNullFloat64(p.FonMinus20, 2)
	case colSensMinus20:
		text = fmtNullFloat64(p.SensMinus20, 2)
	case colFon50:
		text = fmtNullFloat64(p.Fon50, 2)
	case colSens50:
		text = fmtNullFloat64(p.Sens50, 2)
	case colGas24:
		text = fmtNullFloat64(p.I24, 2)
	case colGas35:
		text = fmtNullFloat64(p.I35, 2)
	case colGas26:
		text = fmtNullFloat64(p.I26, 2)
	case colGas17:
		text = fmtNullFloat64(p.I17, 2)
	case colNei:
		text = fmtNullFloat64(p.In, 2)
	case colCalcKSens20:
		text = fmtNullFloat64(p2.CoefficientSensitivity20(), 3)
		if !p2.OkCoefficientSensitivity20() {
			setError()
		}
	case colCalcKSens50:
		text = fmtNullFloat64(p.CoefficientSensitivity50(), 3)
		if !p2.OkCoefficientSensitivity50() {
			setError()
		}

	case colCalcDFon:
		d := p2.DeltaFon20()
		if d.Valid {
			text = fmtFloat64(d.Float64, 3)
		}
		if !p2.OkDeltaFon20() {
			setError()
		}
	case colCalcDT:
		d := p2.DeltaFonTemperature()
		if d.Valid {
			text = fmtFloat64(d.Float64, 3)
		}
		if !p2.OkDeltaFonTemperature() {
			setError()
		}
	case colCalcDNei:
		d := p2.DeltaFonN()
		if d.Valid {
			text = fmtFloat64(d.Float64, 3)
		}
		if !p2.OkDeltaFonN() {
			setError()
		}
	}

	return
}

func (x *ProductsTableModel) StyleCell(c *walk.CellStyle) {
	_, image, textColor, backgroundColor, applyTextColor, applyBackgroundColor := x.Data(c.Row(), c.Col())
	if image != "" {
		c.Image = image
	}
	if applyTextColor {
		c.TextColor = textColor
	}
	if applyBackgroundColor {
		c.BackgroundColor = backgroundColor
	}

}

func (x *ProductsTableModel) Value(row, col int) interface{} {
	str, _, _, _, _, _ := x.Data(row, col)
	return str
}

func (x *ProductsTableModel) Checked(row int) bool {
	p := x.app.products[row]
	return p.Production
}

func (x *ProductsTableModel) SetChecked(row int, checked bool) error {

	if x.app.IsLastParty() {
		x.app.db.UpdateProductProduction(x.app.products[row].ProductID, checked)
		x.app.products[row].Production = checked
	}
	return nil
}

func fmtNullFloat64(x sql.NullFloat64, prec int) string {
	if !x.Valid {
		return ""
	}
	return fmtFloat64(x.Float64, prec)
}

func fmtFloat64(x float64, prec int) string {
	return strconv.FormatFloat(x, 'f', prec, 64)
}
