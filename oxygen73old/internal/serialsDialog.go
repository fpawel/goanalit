package internal

import (
	"fmt"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func (x *AppMainWindow) runSerialsDialog() (int, error) {
	var dlg *walk.Dialog

	var acceptPB, cancelPB *walk.PushButton
	items := make([]*walk.NumberEdit, PartyProductsCount)

	font := Font{
		Family:    "Arial",
		PointSize: 10,
	}

	var children []Widget

	for i := 0; i < PartyProductsCount; i++ {
		children = append(children,
			Composite{
				Layout: Grid{
					Margins:     Margins{30, 0, 10, 0},
					SpacingZero: true,
				},
				Children: []Widget{
					Label{
						Font: Font{
							Family:    "Arial",
							PointSize: 10,
							Bold:      true,
							Italic:    true,
						},
						Text:    fmt.Sprintf("%d", i+1),
						MaxSize: Size{15, 0},
						MinSize: Size{15, 0},
					},
				},
			},

			NumberEdit{
				AssignTo: &items[i],
				Font:     font,
				Value:    float64(x.model.party.serials[i]),
				MinSize:  Size{50, 0},
			},
		)
	}

	return Dialog{
		AssignTo:      &dlg,
		Title:         "Ввод заводских номеров ячеек партии",
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		MinSize:       Size{300, 350},
		Layout:        HBox{},

		Children: []Widget{
			ScrollView{
				HorizontalFixed: true,
				Font:            font,

				Layout: Grid{
					Columns:     10,
					MarginsZero: true,
					SpacingZero: true,
				},
				Children: children,
			},
			ScrollView{
				HorizontalFixed: true,
				Layout:          VBox{},
				Children: []Widget{
					PushButton{
						AssignTo: &acceptPB,
						Text:     "OK",
						OnClicked: func() {
							for i := 0; i < PartyProductsCount; i++ {
								x.model.party.serials[i] = uint(items[i].Value())
							}
							dlg.Accept()
						},
					},
					PushButton{
						AssignTo:  &cancelPB,
						Text:      "Cancel",
						OnClicked: func() { dlg.Cancel() },
					},
				},
			},
		},
	}.Run(x.MainWindow)
}
