package main

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"log"
	"sync/atomic"
)

const (
	FontPointSize = 10
)

func (x *AppMainWindow) Markup() MainWindow {

	var productTableColumns []TableViewColumn
	for _, title := range productColumns {
		productTableColumns = append(productTableColumns, TableViewColumn{
			Title: title,
			Width: 90,
		})
	}

	svWork := ScrollView{
		AssignTo:      &x.svWork,
		VerticalFixed: true,
		Layout:        HBox{SpacingZero: false, MarginsZero: false},
		Children: []Widget{
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
								AssignTo: &x.lblWork,
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
				Image:          Png16RightArrow,
				Text:           "Продолжить",
				AssignTo:       &x.btnCancelDelay,
				ToolTipText:    "Прервать задержку и продолжить выполнение текущей операции",
				OnClicked: func() {
					atomic.AddInt32(&x.app.cancellationDelay, 1)
				},
			},
			PushButton{
				Font:           Font{PointSize: FontPointSize},
				ImageAboveText: true,
				Image:          Png16Cancel,
				Text:           "Прервать",
				AssignTo:       &x.btnCancel,
				ToolTipText:    "Прервать выполнение текущей операции",
				OnClicked: func() {
					atomic.AddInt32(&x.app.cancellationDelay, 1)
				},
			},
		},
	}

	var actionPartiesTreeVisibility *walk.Action

	return MainWindow{
		Title:      x.app.ProductName(),
		Background: SolidColorBrush{walk.RGB(255, 255, 255)},
		Name:       "MainWindow",
		Size:       Size{800, 600},
		Layout:     VBox{},
		AssignTo:   &x.MainWindow,
		Icon:       "assets/ico/app.ico",
		MenuItems: []MenuItem{
			Menu{
				Text: "Файл",
				Items: []MenuItem{
					Separator{},
					Action{
						Text: "Настройки",
					},
				},
			},
			Menu{
				Text: "Помощь",
				Items: []MenuItem{
					Action{
						Text: "О программе",
						OnTriggered: func() {
							log.Panic("ops!")
						},
					},
				},
			},
		},
		Children: []Widget{
			svWork,

			ScrollView{
				Layout:        HBox{MarginsZero: true, Spacing: 3},
				VerticalFixed: true,
				Children: []Widget{

					SplitButton{
						Text:  "Архив",
						Font:  Font{PointSize: 12},
						Image: "assets/png16/db.png",
						MenuItems: []MenuItem{

							Action{
								Image:    "assets/png16/tree.png",
								Text:     "Показать",
								AssignTo: &actionPartiesTreeVisibility,
								OnTriggered: func() {
									f := !x.tvParties.Visible()
									x.tvParties.SetVisible(f)
									s := "Показать"
									if f {
										s = "Скрыть"
									}
									if err := actionPartiesTreeVisibility.SetText(s); err != nil {
										panic(err)
									}
								},
							},
							Action{
								Image:       "assets/png16/search.png",
								Text:        "Поиск ЭХЯ",
								OnTriggered: x.app.ExecuteFindProductBySerialDialog,
							},
						},
					},

					SplitButton{
						Text:  "Партия",
						Font:  Font{PointSize: 12},
						Image: "assets/png16/folder1.png",
						MenuItems: []MenuItem{

							Action{
								Image:       "assets/png16/add-new.png",
								Text:        "Создать новую",
								OnTriggered: x.app.ExecuteNewPartyDialog,
							},

							Action{
								Text:        "Ввод данных",
								Image:       "assets/png16/edit.png",
								OnTriggered: x.app.ExecuteEditPartyDialog,
							},
						},
					},

					Label{
						AssignTo:  &x.lblParty,
						Font:      Font{PointSize: 12},
						TextColor: walk.RGB(0, 0, 200),
					},
					ScrollView{
						Layout:        HBox{MarginsZero: true, SpacingZero: true},
						VerticalFixed: true,
						Children: []Widget{
							Label{},
						},
					},

					ToolButton{
						Image:       "assets/png16/close.png",
						AssignTo:    &x.btnCloseClickedProduct,
						ToolTipText: "Закрыть ЭХЯ",
						OnClicked:   x.HideClickedProduct,
					},
				},
			},

			Composite{
				Font:   Font{PointSize: FontPointSize},
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					TreeView{
						MaxSize:              Size{150, 0},
						MinSize:              Size{150, 0},
						AssignTo:             &x.tvParties,
						Model:                x.app.partiesTreeViewModel,
						OnItemActivated:      x.TreeViewPartiesItemActivated,
						OnCurrentItemChanged: x.TreeViewPartiesItemChanged,
						ContextMenuItems: []MenuItem{
							Action{
								Text:        "Удалить партию",
								AssignTo:    &x.actionDeleteParty,
								OnTriggered: x.app.ExecuteDeletePartyDialog,
							},
						},
					},

					Composite{
						Layout: HBox{MarginsZero: true},

						Children: []Widget{
							TableView{
								AssignTo:              &x.tblProducts,
								AlternatingRowBGColor: walk.RGB(239, 239, 239),
								CheckBoxes:            true,
								ColumnsOrderable:      false,
								MultiSelection:        false,
								Columns:               productTableColumns,
								Model:                 x.app.productsTableModel,
								LastColumnStretched:   true,
								OnItemActivated:       x.TableProductsItemActivated,
								Font:                  Font{PointSize: 12},
							},

							ImageView{
								Background:    SolidColorBrush{walk.RGB(255, 255, 255)},
								AssignTo:      &x.ivChart,
								Mode:          ImageViewModeShrink,
								OnSizeChanged: x.invalidateChart,
							},
						},
					},
				},
			},
		},
	}
}
