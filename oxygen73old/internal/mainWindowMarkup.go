package internal

import (
	"fmt"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func newMainwindow(mw *AppMainWindow) MainWindow {

	var columns []TableViewColumn
	for i := 0; i < 5; i++ {
		columns = append(columns,
			TableViewColumn{
				Title: fmt.Sprintf("%02d", (i+1)*10),
				Width: 30,
			},
			TableViewColumn{
				Width: 60,
			},
			TableViewColumn{
				Width:     70,
				Precision: 3,
			},
		)
	}

	return MainWindow{
		Title:    "Лаборатория №73. Датчик кислорода.",
		Name:     "MainWindow",
		Size:     Size{800, 600},
		Layout:   VBox{MarginsZero: true, SpacingZero: true},
		AssignTo: &mw.MainWindow,

		MenuItems: []MenuItem{
			Menu{
				Text: "&График",
				Items: []MenuItem{
					Action{
						AssignTo:    &mw.browsePartiesAction,
						Text:        "&Открыть",
						Image:       "./img/open.png",
						OnTriggered: mw.browsePartiesAction_Triggered,
					},

					Separator{},
					Action{
						Text:        "Выход",
						OnTriggered: func() { mw.Close() },
					},
				},
			},
			Menu{
				Text: "&Настройки",
				Items: []MenuItem{
					Action{
						AssignTo:    &mw.editSerialsAction,
						Text:        "&Ввод зав.№",
						Image:       "./img/edit.png",
						OnTriggered: mw.editSerialsAction_Triggered,
					},
				},
			},
			Menu{
				Text: "&Помощь",
				Items: []MenuItem{
					Action{
						Text:        "О программе",
						OnTriggered: mw.aboutAction_Triggered,
					},
				},
			},
		},

		Children: []Widget{
			ScrollView{
				Layout: HBox{},
				//HorizontalFixed:true,
				VerticalFixed: true,
				Children: []Widget{
					Label{Text: "СОМ порт"},
					ComboBox{
						Name:     "ComboBoxSerialPort",
						AssignTo: &mw.comboBoxPort,
						OnMouseDown: func(x, y int, button walk.MouseButton) {
							mw.refreshSerials()
						},
					},
					Label{
						AssignTo: &mw.labelPortConn,
						Text:     "Связь устанавливается...",
						Font: Font{
							PointSize: 10,
						},
					},
					Label{Text: " Партия:"},
					Label{
						AssignTo: &mw.labelPartyCreatedAt,
						Font: Font{
							Family:    "Arial",
							Italic:    true,
							PointSize: 10,
						},
					},
					Label{Text: "  Температура, \"C"},
					Label{
						AssignTo: &mw.labelTemperature,
						Text:     "???",
						Font: Font{
							Family:    "Arial",
							PointSize: 12,
						},
					},
					Label{Text: "  Давление, мм.рт.ст."},
					Label{
						AssignTo: &mw.labelPressure,
						Text:     "???",
						Font: Font{
							Family:    "Arial",
							PointSize: 12,
						},
					},
				},
			},
			TableView{
				AssignTo:              &mw.tableViewParty,
				AlternatingRowBGColor: walk.RGB(239, 239, 239),
				CheckBoxes:            false,
				ColumnsOrderable:      false,
				MultiSelection:        true,
				Font: Font{
					Family:    "Arial",
					PointSize: 10,
				},
				Columns:   columns,
				Model:     &mw.model.party,
				StyleCell: mw.StyleCell,
			},
			Composite{
				AssignTo: &mw.panelOpenChart,
				Visible:  false,
				Layout:   HBox{},
				Children: []Widget{
					Label{
						Text: "Загрузка графиков...",
					},
					ProgressBar{
						AssignTo: &mw.progressOpenChart,
					},
				},
			},
		},
	}
}
