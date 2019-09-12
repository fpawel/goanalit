package internal

import (
	"fmt"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

func (x *AppMainWindow) browsePartiesAction_Triggered() {
	var dlg *walk.Dialog
	var acceptPB, cancelPB, delPB *walk.PushButton
	var treeView *walk.TreeView
	m := NewPartiesTreeViewModel(x.model.GetSavedCharts(), x.model.party.chart.CreatedAt, x.model.GetPartyUpdatedAt)
	err := Dialog{
		AssignTo:      &dlg,
		Title:         "Обзор партий",
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		MinSize:       Size{400, 300},
		Layout:        HBox{},
		Children: []Widget{
			TreeView{
				Model:    m,
				AssignTo: &treeView,
				Font: Font{
					Family:    "Arial",
					PointSize: 12,
				},
				OnCurrentItemChanged: func() {
					m.selected = treeView.CurrentItem().(*PartiesNode)
				},
			},
			ScrollView{
				HorizontalFixed: true,
				Layout:          VBox{},
				Children: []Widget{
					PushButton{
						AssignTo: &acceptPB,
						Text:     "Открыть",
						OnClicked: func() {
							if m.selected != nil {
								switch m.selected.nodeType {
								case tnodeParty, tnodeActiveParty:
									go x.model.OpenChart(m.selected.party)
								case tnodeDay:

								}

								dlg.Accept()
							}
						},
					},
					PushButton{
						AssignTo:  &cancelPB,
						Text:      "Отмена",
						OnClicked: func() { dlg.Cancel() },
					},
					Composite{
						Layout: HBox{
							Margins: Margins{Top: 20, Bottom: 20},
						},
						Children: []Widget{
							PushButton{
								Text:     "Удалить",
								AssignTo: &delPB,
								OnClicked: func() {
									if m.selected == nil {
										return
									}
									if m.selected.nodeType == tnodeParty {
										party := m.selected.party
										message := fmt.Sprintf("Подтвердите необходимость удаления партии  %s",
											party.CreatedAt.Format("2006 01 02 03:04:05"))
										did := walk.MsgBox(x, "Удаление данных",
											message,
											walk.MsgBoxOKCancel|walk.MsgBoxIconWarning)
										if did == win.IDOK {

										}

									}
								},
							},
						},
					},
				},
			},
		},
	}.Create(x.MainWindow)
	if err != nil {
		logger.Fatal(err)
	}

	dlg.Run()
}
