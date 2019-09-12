package dialogEditProduct

import (
	"database/sql"
	"fmt"
	"github.com/fpawel/eccco73/internal/eccco73"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	"log"
	"strings"
)

type ProductTypeModel struct {
	Name          string
	ProductTypeID sql.NullInt64
}

func Execute(owner walk.Form, party eccco73.Party, product eccco73.Product, productTypes []ProductTypeModel) (eccco73.Product, bool) {
	dlg := &dialog{
		originalProduct: product,
		inputProduct:    product,
		party:           party,
		productTypes:    productTypes,
	}
	r, err := dlg.markup().Run(owner)
	if err != nil {
		log.Panic(err)
	}
	return dlg.inputProduct, r == win.IDOK
}

type dialog struct {
	dlg          *walk.Dialog
	acceptPB     *walk.PushButton
	productTypes []ProductTypeModel
	party        eccco73.Party
	originalProduct,
	inputProduct eccco73.Product
}

func (x *dialog) validate() {
	x.acceptPB.SetEnabled(x.originalProduct.Note != x.inputProduct.Note ||
		x.originalProduct.Serial != x.inputProduct.Serial ||
		x.originalProduct.ProductTypeID != x.inputProduct.ProductTypeID)
}

func (x *dialog) markup() Dialog {

	var edSerial *walk.NumberEdit
	var cbProductType *walk.ComboBox
	var edNote *walk.LineEdit
	var cancelPB *walk.PushButton
	var cbProductTypeIndex int

	for i, pt := range x.productTypes {
		if pt.ProductTypeID == x.originalProduct.ProductTypeID {
			cbProductTypeIndex = i
		}
	}

	return Dialog{
		Icon:          "assets/ico/settings.ico",
		AssignTo:      &x.dlg,
		Title:         "Изменение ЭХЯ " + eccco73.FormatOrder8(x.originalProduct.Order),
		DefaultButton: &x.acceptPB,
		CancelButton:  &cancelPB,
		FixedSize:     true,
		Font:          Font{PointSize: 10},

		Layout: Grid{Columns: 2},
		Children: []Widget{
			Label{Text: "Заводской номер:"},
			NumberEdit{
				MinSize:  Size{180, 0},
				AssignTo: &edSerial,
				Value:    float64(x.originalProduct.Serial),
				OnValueChanged: func() {
					x.inputProduct.Serial = int64(edSerial.Value())
					x.validate()
				},
			},
			Label{Text: "Исполнение:"},
			ComboBox{
				MinSize:       Size{180, 0},
				AssignTo:      &cbProductType,
				Model:         x.productTypes,
				DisplayMember: "Name",
				CurrentIndex:  cbProductTypeIndex,
				OnCurrentIndexChanged: func() {
					x.inputProduct.ProductTypeID = x.productTypes[cbProductType.CurrentIndex()].ProductTypeID
					x.validate()
				},
			},
			ScrollView{
				Layout:        HBox{MarginsZero: true, SpacingZero: true},
				VerticalFixed: true,
				Children: []Widget{
					Label{Text: "Примечание:"},
				},
			},
			LineEdit{
				AssignTo:   &edNote,
				ColumnSpan: 2,
				Text:       x.originalProduct.Note.String,
				OnTextChanged: func() {
					s := strings.TrimSpace(edNote.Text())
					x.inputProduct.Note.String = s
					x.inputProduct.Note.Valid = len(s) > 0
					if !x.originalProduct.Note.Valid && !x.inputProduct.Note.Valid {
						x.inputProduct.Note.String = x.originalProduct.Note.String
					}
					x.validate()
				},
			},
			Composite{
				ColumnSpan: 2,
				Layout:     HBox{MarginsZero: true},
				Children: []Widget{
					PushButton{
						Text:     "Ок",
						AssignTo: &x.acceptPB,
						Image:    "assets/png16/checkmark.png",
						OnClicked: func() {
							x.dlg.Accept()
						},
						Enabled: false,
					},
					PushButton{
						Text:     "Отмена",
						AssignTo: &cancelPB,
						Image:    "assets/png16/cancel16.png",
						OnClicked: func() {
							fmt.Println(x.dlg.Size())
							x.dlg.Cancel()
						},
					},
				},
			},
		},
	}
}
