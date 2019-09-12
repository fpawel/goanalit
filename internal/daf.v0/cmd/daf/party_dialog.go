package main

import (
	"github.com/fpawel/daf.v0/internal/data"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"math"
)

func runPartyDialog(owner walk.Form) {
	var (
		dlg           *walk.Dialog
		cbGas, cbType *walk.ComboBox
		btn           *walk.PushButton
		saveOnEdit    bool
		party         = data.GetLastParty()
	)

	saveParty := func() {
		if saveOnEdit {
			if err := data.DBProducts.Save(party); err != nil {
				walk.MsgBox(dlg, "Ошибка данных", err.Error(), walk.MsgBoxIconError|walk.MsgBoxOK)
			}
		}
	}

	type NameCode struct {
		Name string
		Code int
	}

	dafTypes := []NameCode{
		{"ДАФ-М-01", 1},
		{"ДАФ-М-05X", 6},
		{"ДАФ-М-06TPX", 9},
		{"ДАФ-М-08X", 80},
		{"ДАФ-М-08TPX", 85},
	}

	dafGases := []string{
		"ацетон С₃Н₆",
		"гексан C₆H₁₄",
		"бензол С₆Н₆",
		"стирол С₈Н₈",
		"толуол С₆Н₅СН₃",
		"фенол С₆Н₆О",
		"этанол С₂Н₅ОН",
		"циклогексан С₆Н₁₂",
	}

	widgets := []Widget{
		Label{Text: "Исполнение:", TextAlignment: AlignFar},
		ComboBox{
			Model:         dafTypes,
			AssignTo:      &cbType,
			DisplayMember: "Name",
			CurrentIndex: func() int {
				for i, x := range dafTypes {
					if x.Code == party.Type {
						return i
					}
				}
				return -1
			}(),
			OnCurrentIndexChanged: func() {
				party.Type = dafTypes[cbType.CurrentIndex()].Code
				saveParty()
			},
		},

		Label{Text: "Компонент:", TextAlignment: AlignFar},
		ComboBox{
			Model:    dafGases,
			AssignTo: &cbGas,
			CurrentIndex: func() int {
				for i, x := range dafGases {
					if x == party.Component {
						return i
					}
				}
				return -1
			}(),
			OnCurrentIndexChanged: func() {
				party.Component = dafGases[cbGas.CurrentIndex()]
				saveParty()
			},
		},
	}

	nePartyField := func(what string, pValue *float64) {
		var ne *walk.NumberEdit
		widgets = append(widgets,
			Label{Text: what, TextAlignment: AlignFar},
			NumberEdit{
				Value:    *pValue,
				AssignTo: &ne,
				MinValue: 0,
				MaxValue: math.MaxFloat64,
				OnValueChanged: func() {
					*pValue = ne.Value()
					saveParty()
				},
			})
	}

	nePartyField("Диапазон:", &party.Scale)
	nePartyField("Дапазон абс. погр.:", &party.AbsoluteErrorRange)
	nePartyField("Предел абс. погр.:", &party.AbsoluteErrorLimit)
	nePartyField("Предел отн. погр., %:", &party.RelativeErrorLimit)

	nePartyField("ПГС1:", &party.Pgs1)

	nePartyField("ПГС2:", &party.Pgs2)
	nePartyField("ПГС3:", &party.Pgs3)
	nePartyField("ПГС4:", &party.Pgs4)

	nePartyField("Порог 1:", &party.Threshold1Production)
	nePartyField("Порог 2:", &party.Threshold2Production)

	nePartyField("Порог 1, настройка:", &party.Threshold1Test)
	nePartyField("Порог 2, настройка:", &party.Threshold2Test)

	widgets = append(widgets,
		Composite{}, Composite{}, Composite{}, Composite{},

		Composite{}, Composite{}, Composite{},
		PushButton{
			AssignTo: &btn,
			Text:     "Закрыть",
			OnClicked: func() {
				dlg.Accept()
			},
		})

	dialog := Dialog{
		Title:         "Параметры партии",
		Font:          Font{PointSize: 12, Family: "Segoe UI"},
		AssignTo:      &dlg,
		Layout:        Grid{Columns: 4},
		DefaultButton: &btn,
		CancelButton:  &btn,
		Children:      widgets,
	}
	if err := dialog.Create(owner); err != nil {
		walk.MsgBox(owner, "Параметры партии", err.Error(), walk.MsgBoxIconError|walk.MsgBoxOK)
		return
	}
	saveOnEdit = true
	dlg.Run()
}
