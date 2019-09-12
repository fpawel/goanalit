package dialogEditParty

import (
	"github.com/fpawel/eccco73/internal/eccco73"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	"log"
	"strings"
)

type dialog struct {
	acceptPB, cancelPB *walk.PushButton
	dlg                *walk.Dialog
	party              eccco73.Party
	productTypes       []eccco73.ProductType
}

func Execute(owner walk.Form, party eccco73.Party, productTypes []eccco73.ProductType) (eccco73.Party, bool) {
	dlg := &dialog{
		productTypes: productTypes,
		party:        party,
	}
	r := dlg.run(owner)
	return dlg.party, r
}

func (x *dialog) run(owner walk.Form) bool {

	var edNote *walk.LineEdit
	var edGas1, edGas2, edGas3 *walk.NumberEdit
	var cbProductTypeIndex *walk.ComboBox

	dlg := Dialog{
		Icon:          "assets/ico/settings.ico",
		AssignTo:      &x.dlg,
		Title:         "Изменение партии",
		DefaultButton: &x.acceptPB,
		CancelButton:  &x.cancelPB,

		FixedSize: true,
		Font:      Font{PointSize: 10},

		Layout: Grid{Columns: 2},
		Children: []Widget{

			Label{Text: "Исполнение:"},
			ComboBox{
				MinSize:       Size{150, 0},
				AssignTo:      &cbProductTypeIndex,
				Model:         x.productTypes,
				DisplayMember: "Name",
				OnCurrentIndexChanged: func() {
					x.party.ProductTypeID = x.productTypes[cbProductTypeIndex.CurrentIndex()].ProductTypeID
				},
				CurrentIndex: 0,
			},
			Label{Text: "ПГС1:"},
			NumberEdit{
				MinSize:        Size{150, 0},
				MinValue:       0,
				MaxValue:       1000,
				Value:          x.party.Gas1,
				Decimals:       1,
				AssignTo:       &edGas1,
				OnValueChanged: func() { x.party.Gas1 = edGas1.Value() },
			},
			Label{Text: "ПГС2:"},
			NumberEdit{
				MinSize:        Size{150, 0},
				MinValue:       0,
				MaxValue:       1000,
				Decimals:       1,
				Value:          x.party.Gas2,
				AssignTo:       &edGas2,
				OnValueChanged: func() { x.party.Gas2 = edGas2.Value() },
			},
			Label{Text: "ПГС3:"},
			NumberEdit{
				MinSize:        Size{150, 0},
				MinValue:       0,
				MaxValue:       1000,
				Decimals:       1,
				Value:          x.party.Gas3,
				AssignTo:       &edGas3,
				OnValueChanged: func() { x.party.Gas3 = edGas3.Value() },
			},
			ScrollView{
				Layout:        HBox{MarginsZero: true, SpacingZero: true},
				VerticalFixed: true,
				Children: []Widget{
					Label{Text: "Примечание:"},
				},
			},

			LineEdit{
				ColumnSpan: 2,
				AssignTo:   &edNote,
				Text:       x.party.Note.String,
				OnTextChanged: func() {
					x.party.Note.String = strings.TrimSpace(edNote.Text())
					x.party.Note.Valid = len(x.party.Note.String) > 0
				},
			},
			Composite{
				ColumnSpan: 2,
				Layout:     HBox{},
				Children: []Widget{
					PushButton{
						Text:     "Ок",
						AssignTo: &x.acceptPB,
						Image:    "assets/png16/checkmark.png",
						OnClicked: func() {
							x.dlg.Accept()
						},
					},
					PushButton{
						Text:     "Отмена",
						AssignTo: &x.cancelPB,
						Image:    "assets/png16/cancel16.png",
						OnClicked: func() {
							x.dlg.Cancel()
						},
					},
				},
			},
		},
	}

	if err := dlg.Create(owner); err != nil {
		log.Panic(err)
	}
	for i, pt := range x.productTypes {
		if pt.ProductTypeID == x.party.ProductTypeID {
			if err := cbProductTypeIndex.SetCurrentIndex(i); err != nil {
				log.Panic(err)
			}
			break
		}
	}

	return x.dlg.Run() == win.IDOK
}
