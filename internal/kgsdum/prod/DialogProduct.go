package main

import (
	wask62 "github.com/fpawel/goanalit/internal/kgsdum/wask6"
	"github.com/fpawel/gutils/walkUtils"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

func ExecuteSetAddrDialog(owner walk.Form) (byte, bool) {
	var dlg *walk.Dialog
	var ed *walk.NumberEdit
	var btnOk, btnCancel *walk.PushButton

	dlgResult, err := Dialog{
		AssignTo:      &dlg,
		Title:         "Установка адреса",
		Layout:        VBox{},
		FixedSize:     true,
		DefaultButton: &btnOk,
		CancelButton:  &btnCancel,
		Font:          Font{PointSize: 14},

		Children: []Widget{
			Label{Text: "Адрес:"},
			NumberEdit{
				AssignTo: &ed,
				MinValue: 1,
				MaxValue: 127,
				Decimals: 0,
			},

			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{
						Text:     "Ок",
						AssignTo: &btnOk,
						Image:    AssetImage("assets/png16/checkmark.png"),
						OnClicked: func() {
							dlg.Accept()
						},
					},
					PushButton{
						Text:     "Отмена",
						AssignTo: &btnCancel,
						Image:    AssetImage("assets/png16/cancel16.png"),
						OnClicked: func() {
							dlg.Cancel()
						},
					},
				},
			},
		},
	}.Run(owner)
	check(err)
	return byte(ed.Value()), dlgResult == win.IDOK
}

func (x *App) ExecuteSendCommandDialog() (wask62.ValueAddr, float32, bool) {
	var dlg *walk.Dialog
	var edCmdCode, edArgValue *walk.NumberEdit
	var btnOk, btnCancel *walk.PushButton

	dlgResult, err := Dialog{
		AssignTo:      &dlg,
		Title:         "Отправка команды",
		Layout:        VBox{},
		FixedSize:     true,
		DefaultButton: &btnOk,
		CancelButton:  &btnCancel,
		Font:          Font{PointSize: 14},

		Children: []Widget{

			walkUtils.LeftAlignedTitleLabel("Код команды:"),
			NumberEdit{
				AssignTo: &edCmdCode,
				MinValue: 0,
				MaxValue: 0xFFFF,
				Decimals: 0,
				Value:    x.config.UserInput.SendCommand.Code,
				OnValueChanged: func() {
					x.config.UserInput.SendCommand.Code = edCmdCode.Value()
				},
			},

			walkUtils.LeftAlignedTitleLabel("Значение аргумента:"),
			NumberEdit{
				AssignTo: &edArgValue,
				MinValue: -999999.999999,
				MaxValue: 999999.999999,
				Decimals: 6,
				Value:    x.config.UserInput.SendCommand.Value,
				OnValueChanged: func() {
					x.config.UserInput.SendCommand.Value = edArgValue.Value()
				},
			},

			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{
						Text:     "Ок",
						AssignTo: &btnOk,
						Image:    AssetImage("assets/png16/checkmark.png"),
						OnClicked: func() {
							dlg.Accept()
						},
					},
					PushButton{
						Text:     "Отмена",
						AssignTo: &btnCancel,
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
	return wask62.ValueAddr(edCmdCode.Value()), float32(edArgValue.Value()), dlgResult == win.IDOK
}
