package productsView

import (
	"fmt"
	"github.com/fpawel/goanalit/internal/kgsdumal/kgsdum/products"
	"github.com/lxn/walk"
	//. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

func DeleteNode(node Node, tx products.Tx) {
	promptMessage := fmt.Sprintf("Подтвердите необходимость удаления\n%s", What())
	if walk.MsgBox(walk.App().ActiveForm(), "Удаление данных", promptMessage, walk.MsgBoxOKCancel|walk.MsgBoxIconWarning) != win.IDOK {
		return
	}
	tx.DeleteParties(func(partyTime products.PartyTime) bool {
		return ContainsParty(partyTime)
	})
}

/*
func executeDeleteCurrentPartyDialog(strCurrentParty string) int {

	var dlg *walk.Dialog
	count := 20
	var btnOk, btnCancel *walk.PushButton
	var ed *walk.LineEdit

	result, err := (Dialog{
		AssignTo:      &dlg,
		CancelButton:  &btnCancel,
		DefaultButton: &btnOk,

		Title:   "Количество приборов",
		MinSize: Size{300, 220},
		MaxSize: Size{300, 220},
		Layout:  HBox{},
		Font:    Font{PointSize: 12},
		Children: []Widget{
			ScrollView{
				HorizontalFixed: true,
				Layout:          VBox{},
				Children: []Widget{
					Label{Text: "Текущая партия будет удалена:"},
					Label{Text: strCurrentParty, Font: Font{Bold: true}},
					Label{Text: ""},
					Label{Text: ""},
					Label{Text: "Введите количество приборов"},
					Label{Text: "в новой партии"},
					LineEdit{
						Text:     "20",
						AssignTo: &ed,
						OnTextChanged: func() {
							n, err := strconv.Atoi(ed.Text())
							if err != nil || n < 1 || n > 20 {
								btnOk.SetEnabled(false)
								return
							}
							count = n
							btnOk.SetEnabled(true)

						},
					},
				},
			},
			ScrollView{
				HorizontalFixed: true,
				Layout:          VBox{},
				Children: []Widget{
					PushButton{
						AssignTo: &btnOk,
						Text:     "Ok",
						OnClicked: func() {
							dlg.Accept()
						},
					},
					PushButton{
						Text:      "Отмена",
						AssignTo:  &btnCancel,
						OnClicked: func() { dlg.Cancel() },
					},
				},
			},
		},
	}).run(walk.App().ActiveForm())
	if err != nil {
		log.Fatal(err)
	}
	if result == win.IDOK {
		return count
	}
	return 0
}
*/
