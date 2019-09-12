package main

import (
	productsView2 "github.com/fpawel/goanalit/internal/kgsdum/productsView"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"sync/atomic"
	"time"
)

type Alignment int

const (
	FontPointSize = 10
)

const (
	AlignTop Alignment = iota
	AlignBottom
)

func (x *AppMainWindow) Markup() MainWindow {

	partyColumns := []TableViewColumn{
		{Title: "БО", Width: 120},
	}

	tableWorks := TableView{
		Model:            x.app.tableMainWorksModel,
		MaxSize:          Size{0, 190},
		MinSize:          Size{0, 190},
		AssignTo:         &x.tblWorks,
		CheckBoxes:       true,
		ColumnsOrderable: false,
		MultiSelection:   false,
		Font:             Font{PointSize: FontPointSize},
		Columns: []TableViewColumn{
			{Title: "Действие", Width: 200},
			{Title: "Статус", Width: 400},
		},
	}

	tableProducts := TableView{
		AssignTo:              &x.tblProducts,
		AlternatingRowBGColor: walk.RGB(239, 239, 239),
		CheckBoxes:            true,
		ColumnsOrderable:      false,
		MultiSelection:        true,
		Font: Font{
			Family:    "Arial",
			PointSize: 10,
		},
		Columns: partyColumns,
		Model:   x.app.tableProductsModel,
	}

	tableLog := TableView{
		AssignTo:              &x.tblLogs,
		AlternatingRowBGColor: walk.RGB(239, 239, 239),
		CheckBoxes:            false,
		ColumnsOrderable:      false,
		MultiSelection:        true,
		Font:                  Font{PointSize: FontPointSize},
		Model:                 x.app.tableLogsModel,

		Columns: productsView2.TableLogsColumns(),
	}

	topInfoPanel := ScrollView{
		VerticalFixed: true,

		Layout: HBox{SpacingZero: false, MarginsZero: false},
		Children: []Widget{
			SplitButton{
				Font:           Font{PointSize: FontPointSize},
				AssignTo:       &x.btnRun,
				ImageAboveText: true,
				Image:          AssetImage("assets/png16/right-arrow.png"),
				Text:           "Пуск",
				ToolTipText:    "Запуск выполнения сценария настройки приборов",
				MenuItems: []MenuItem{
					Action{
						Text:        "Опрос",
						OnTriggered: x.app.RunManualSurvey,
					},
					Action{
						Text:        "Установка адреса",
						OnTriggered: x.app.RunSetAddr,
					},
					Action{
						Text:        "Записать коэффициенты",
						OnTriggered: x.app.WriteCoefficients,
					},
					Action{
						Text:        "Считать коэффициенты",
						OnTriggered: x.app.ReadCoefficients,
					},
					Action{
						Text:        "Отправка команды",
						OnTriggered: x.app.RunSendCommand,
					},
					Separator{},

					Action{
						Text: "Test delay",
						OnTriggered: func() {
							x.InvalidateWorkRunning(true)
							go func() {
								x.app.Delay("Test delay", time.Minute, func() {
									time.Sleep(time.Second / 3)
								})
								x.Synchronize(func() {
									x.InvalidateWorkRunning(false)
								})
							}()
						},
					},
				},
				OnClicked: x.app.RunMainWork,
			},
			PushButton{
				Font:           Font{PointSize: FontPointSize},
				ImageAboveText: true,
				Image:          AssetImage("assets/png16/cancel.png"),
				Text:           "Прервать",
				AssignTo:       &x.btnCancel,
				ToolTipText:    "Прервать выполнение текущей операции",
				OnClicked: func() {
					x.app.ports.Cancel()
					atomic.AddInt32(&x.app.cancellationDelay, 1)
				},
			},
			Composite{
				Layout: VBox{
					SpacingZero: true, MarginsZero: true,
				},
				Children: []Widget{
					ScrollView{
						Layout:        HBox{MarginsZero: true, SpacingZero: true},
						VerticalFixed: true,
						Children: []Widget{
							Label{
								AssignTo: &x.lblWorkInfo,
								Font:     Font{PointSize: FontPointSize},
							},
						},
					},
					TextEdit{
						AssignTo: &x.lblWorkMessage,
						ReadOnly: true,
						Font:     Font{PointSize: FontPointSize},
					},
				},
			},
			ProgressBar{
				AssignTo: &x.progressBar,
			},
			PushButton{
				Font:           Font{PointSize: FontPointSize},
				ImageAboveText: true,
				Image:          AssetImage("assets/png16/right-arrow.png"),
				Text:           "Продолжить",
				AssignTo:       &x.btnCancelDelay,
				ToolTipText:    "Прервать задержку и продолжить выполнение текущей операции",
				OnClicked: func() {
					atomic.AddInt32(&x.app.cancellationDelay, 1)
				},
			},
		},
	}

	return MainWindow{
		Title:    x.app.ProductName(),
		Name:     "MainWindow",
		Size:     Size{800, 600},
		Layout:   HBox{MarginsZero: true, SpacingZero: true},
		AssignTo: &MainWindow,
		Icon:     NewIconFromResourceId(IconAppID),
		MenuItems: []MenuItem{
			Menu{
				Text: "Файл",
				Items: []MenuItem{
					Action{
						Text:        "Новая партия",
						OnTriggered: x.app.ExecuteNewPartyDialog,
					},
					Action{
						Text:        "Архив",
						OnTriggered: x.app.ExecuteDialogArchive,
					},
					Action{
						Text:        "Ввод номеров",
						OnTriggered: x.app.ExecutePartyNumbersDialog,
					},
					Separator{},
					Action{
						Text:        "Настройки",
						OnTriggered: x.app.ExecuteAppConfigDialog,
					},
				},
			},
			Menu{
				Text: "Помощь",
				Items: []MenuItem{
					Action{
						Text: "О программе",
					},
				},
			},
		},
		Children: []Widget{
			Composite{
				Layout: VBox{},
				Children: []Widget{
					topInfoPanel,
					Composite{
						Layout: HBox{},
						Children: []Widget{
							Composite{
								Layout: VBox{},
								Children: []Widget{
									tableWorks,
									tableProducts,
								},
							},
							tableLog,
						},
					},
				},
			},
			ScrollView{
				Name:            "RightPan",
				Layout:          VBox{},
				HorizontalFixed: true,
				Children:        []Widget{},
			},
		},
	}
}
func (x *AppMainWindow) setWidgetHeigh1(w **walk.TableView, h int) {
	t := *w
	t.SetHeight(h)
	t.SetMinMaxSize(walk.Size{0, h}, walk.Size{0, h})
	x.Invalidate()
}

func (x *AppMainWindow) sizableTableMarkup(tableMarkup TableView,
	table **walk.TableView, configHeight *int, align Alignment) ScrollView {

	btnUp := ToolButton{
		MinSize:     Size{16, 16},
		MaxSize:     Size{16, 16},
		Image:       AssetImage("assets/png16/arrow_up2.png"),
		ToolTipText: "Растянуть таблицу",
		OnClicked: func() {
			h := *configHeight
			h += 10
			if h > 500 {
				return
			}
			x.setWidgetHeigh1(table, h)
			*configHeight = (*table).Height()
		},
	}

	btnDown := ToolButton{
		MinSize:     Size{16, 16},
		MaxSize:     Size{16, 16},
		Image:       AssetImage("assets/png16/arrow_down2.png"),
		ToolTipText: "Растянуть таблицу",
		OnClicked: func() {
			h := *configHeight
			h -= 10
			if h < 80 {
				return
			}
			x.setWidgetHeigh1(table, h)
			*configHeight = (*table).Height()
		},
	}

	var xs []Widget

	switch align {
	case AlignTop:
		xs = []Widget{btnUp, btnDown}
	case AlignBottom:
		xs = []Widget{ToolBar{Orientation: Vertical}, btnUp, btnDown}
	}

	return ScrollView{
		Layout:        HBox{MarginsZero: true, SpacingZero: true},
		VerticalFixed: true,
		Children: []Widget{
			tableMarkup,
			ScrollView{
				Layout:          VBox{MarginsZero: true, SpacingZero: true},
				HorizontalFixed: true,
				Children:        xs,
			},
		},
	}
}
