package main

import (
	"fmt"
	products2 "github.com/fpawel/goanalit/internal/kgsdum/products"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

func (x *App) ExecuteNewPartyDialog() {
	var acceptPB, cancelPB *walk.PushButton
	var dlg *walk.Dialog
	var productsCountNE *walk.NumberEdit

	dlgResult, err := Dialog{
		AssignTo:      &dlg,
		Title:         "Новая партия",
		Layout:        VBox{},
		FixedSize:     true,
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		Font:          Font{PointSize: 14},

		Children: []Widget{
			Label{Text: "Количество приборов в новой партии:"},
			NumberEdit{
				AssignTo: &productsCountNE,
				MinValue: 1,
				MaxValue: 30,
				Decimals: 0,
			},

			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{
						Text:     "Ок",
						AssignTo: &acceptPB,
						Image:    AssetImage("assets/png16/checkmark.png"),
						OnClicked: func() {
							dlg.Accept()
						},
					},
					PushButton{
						Text:     "Отмена",
						AssignTo: &cancelPB,
						Image:    AssetImage("assets/png16/cancel16.png"),
						OnClicked: func() {
							dlg.Cancel()
						},
					},
				},
			},
		},
	}.Run(x.mw)
	check(err)

	if dlgResult != win.IDOK {
		return
	}

	n := int(productsCountNE.Value())

	x.db.Update(func(tx products2.Tx) {
		tx.NewParty(n)
	})
	x.ExecutePartyNumbersDialog()

}

func (x *App) ExecutePartyNumbersDialog() {
	var acceptPB, cancelPB *walk.PushButton
	var dlg *walk.Dialog

	xs := []Widget{
		Label{Text: "№"},
		Label{Text: "Адрес"},
		Label{Text: "Серийный №"},
	}
	ps := x.db.Products()
	tmp := make([]products2.ProductInfo, len(ps))
	copy(tmp, ps)

	for i, y := range tmp {
		i := i
		y := y
		var addrNE, serialNE *walk.NumberEdit
		xs = append(xs, []Widget{
			Label{
				Text: fmt.Sprintf("%d", i+1),
			},
			NumberEdit{
				AssignTo: &addrNE,
				MinValue: 0,
				MaxValue: 127,
				Decimals: 0,
				Value:    float64(y.Addr),
				OnValueChanged: func() {
					v := byte(addrNE.Value())
					tmp[i].Addr = v
				},
			},
			NumberEdit{
				AssignTo: &serialNE,
				MinValue: 0,
				MaxValue: 10000000,
				Decimals: 0,
				Value:    float64(y.Serial),
				OnValueChanged: func() {
					v := uint64(serialNE.Value())
					tmp[i].Serial = v
				},
			},
		}...)
	}

	dlgResult, err := Dialog{
		Icon:          NewIconFromResourceId(IconSettingsID),
		AssignTo:      &dlg,
		Title:         "Ввод номеров",
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		MinSize:       Size{300, 500},
		Layout:        HBox{},
		Font:          Font{PointSize: 14},
		FixedSize:     true,

		Children: []Widget{
			ScrollView{
				HorizontalFixed: true,
				Layout: Grid{
					Columns: 3,
				},
				Children: xs,
			},
			ScrollView{
				HorizontalFixed: true,
				Layout:          VBox{},
				Children: []Widget{
					PushButton{
						Text:     "Ок",
						AssignTo: &acceptPB,
						Image:    AssetImage("assets/png16/checkmark.png"),
						OnClicked: func() {
							dlg.Accept()
						},
					},
					PushButton{
						Text:     "Отмена",
						AssignTo: &cancelPB,
						Image:    AssetImage("assets/png16/cancel16.png"),
						OnClicked: func() {
							dlg.Cancel()
						},
					},
				},
			},
		},
	}.Run(x.mw)
	check(err)

	if dlgResult != win.IDOK {
		return
	}

	x.db.Update(func(tx products2.Tx) {
		party := tx.Party()
		for i, p := range party.Products() {
			y := tmp[i]
			p.SetAddr(y.Addr)
			p.SetSerial(y.Serial)
		}
	})

	x.tableProductsModel.PublishRowsReset()
	x.tableLogsModel.PublishRowsReset()

}
