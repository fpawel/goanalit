package dialogNewParty

import (
	"fmt"
	"github.com/fpawel/eccco73/internal/eccco73"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	"log"
	"strconv"
	"strings"
)

func Execute(owner walk.Form, productTypes []eccco73.ProductType) (newParty eccco73.NewParty, ok bool) {
	dlg := &dlg{
		productTypes: productTypes,
	}
	ok = dlg.run(owner)
	newParty = dlg.newParty
	return
}

type dlg struct {
	acceptPB, cancelPB *walk.PushButton
	dlg                *walk.Dialog
	eds                [8 * 12]*walk.LineEdit
	productTypes       []eccco73.ProductType
	newParty           eccco73.NewParty
}

var errorNoInput = fmt.Errorf("NO INPUT")

func (x *dlg) validateSerialAt(index int) error {
	x.newParty.Serials[index] = 0
	s := strings.TrimSpace(x.eds[index].Text())
	if s == "" {
		return errorNoInput
	}
	var v int
	v, err := strconv.Atoi(s)
	if err != nil {
		return err
	}

	if v < 1 {
		return fmt.Errorf("серийный номер должен быть больше нуля")
	}
	for i, ed := range x.eds {
		if i == index {
			continue
		}
		a, e := strconv.Atoi(strings.TrimSpace(ed.Text()))
		if e == nil && a == v {
			return fmt.Errorf("дублирование номера %d: %d.%d и %d.%d", a,
				i/8+1, i%8+1,
				index/8+1, index%8+1,
			)
		}
	}
	x.newParty.Serials[index] = int64(v)
	return nil

}

func (x *dlg) validateInput() {

	valid := true
	hasValid := false
	for index, ed := range x.eds {

		brushColor := walk.RGB(255, 255, 255)
		fontColor := walk.RGB(0, 0, 0)
		toolTipText := fmt.Sprintf("ЭХЯ %d.%d", index/8+1, index%8+1)

		if err := x.validateSerialAt(index); err != nil {
			brushColor = walk.RGB(242, 242, 242)
			if err != errorNoInput {
				toolTipText += ":" + err.Error()
				fontColor = walk.RGB(255, 0, 0)
				valid = false
			}
		} else {
			toolTipText += ": ok"
			hasValid = true
		}

		if ed.ToolTipText() != toolTipText {
			if err := ed.SetToolTipText(toolTipText); err != nil {
				log.Panic(err)
			}
			b, err := walk.NewSolidColorBrush(brushColor)
			if err != nil {
				log.Panic(err)
			}
			ed.SetBackground(b)
			ed.SetTextColor(fontColor)
		}
	}
	x.acceptPB.SetEnabled(valid && hasValid)
}

func (x *dlg) run(owner walk.Form) bool {

	xs := []Widget{
		Label{Text: "№"},
	}
	for i := 0; i < 8; i++ {
		xs = append(xs, Label{Text: strconv.Itoa(i + 1)})
	}
	for i := 0; i < 12; i++ {
		xs = append(xs, Label{
			Text:    strconv.Itoa(i + 1),
			MinSize: Size{0, 20},
			MaxSize: Size{0, 20},
		})
		for j := 0; j < 8; j++ {
			n := i*8 + j
			xs = append(xs, LineEdit{
				AssignTo:      &x.eds[n],
				MinSize:       Size{60, 0},
				MaxSize:       Size{60, 0},
				OnTextChanged: x.validateInput,
				Background:    SolidColorBrush{Color: walk.RGB(242, 242, 242)},
			})
		}
	}

	var edNote *walk.LineEdit
	var edGas1, edGas2, edGas3 *walk.NumberEdit
	var cbProductTypeIndex *walk.ComboBox

	dlg := Dialog{
		Icon:          "assets/ico/settings.ico",
		AssignTo:      &x.dlg,
		Title:         "Ввод номеров",
		DefaultButton: &x.acceptPB,
		CancelButton:  &x.cancelPB,
		MinSize:       Size{530, 600},
		MaxSize:       Size{530, 600},
		Layout:        HBox{},
		Font:          Font{PointSize: 10},
		FixedSize:     true,

		Children: []Widget{
			ScrollView{
				HorizontalFixed: true,
				Layout:          VBox{},
				Children: []Widget{

					Composite{
						Layout: Grid{Columns: 4, MarginsZero: true},
						Children: []Widget{
							Label{Text: "Исполнение:"},
							ComboBox{
								AssignTo:      &cbProductTypeIndex,
								Model:         x.productTypes,
								DisplayMember: "Name",
								OnCurrentIndexChanged: func() {
									x.newParty.ProductType = x.productTypes[cbProductTypeIndex.CurrentIndex()]
								},
								CurrentIndex: 0,
							},
							Label{Text: "ПГС1:"},
							NumberEdit{
								MinValue:       0,
								MaxValue:       1000,
								Decimals:       1,
								AssignTo:       &edGas1,
								OnValueChanged: func() { x.newParty.Gas1 = edGas1.Value() },
							},
							Label{Text: "ПГС2:"},
							NumberEdit{
								MinValue:       0,
								MaxValue:       1000,
								Decimals:       1,
								Value:          float64(50),
								AssignTo:       &edGas2,
								OnValueChanged: func() { x.newParty.Gas2 = edGas2.Value() },
							},
							Label{Text: "ПГС3:"},
							NumberEdit{
								MinValue:       0,
								MaxValue:       1000,
								Decimals:       1,
								Value:          float64(100),
								AssignTo:       &edGas3,
								OnValueChanged: func() { x.newParty.Gas3 = edGas3.Value() },
							},
							Label{Text: "Примечание:"},
							LineEdit{
								ColumnSpan: 4,
								AssignTo:   &edNote,
								OnTextChanged: func() {
									x.newParty.Note.String = strings.TrimSpace(edNote.Text())
									x.newParty.Note.Valid = len(x.newParty.Note.String) > 0
								},
							},
						},
					},

					ScrollView{
						HorizontalFixed: true,
						Layout: Grid{
							MarginsZero: true,
							Columns:     9,
						},
						Children: xs,
					},
				},
			},

			ScrollView{
				HorizontalFixed: true,
				Layout:          VBox{},
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

	res, err := dlg.Run(owner)
	if err != nil {
		log.Panic(err)
	}
	return res == win.IDOK
}
