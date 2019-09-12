package main

import (
	"github.com/fpawel/guartutils/comport"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"time"
)

func (x *App) ExecuteAppConfigDialog() {
	var acceptPB, cancelPB *walk.PushButton

	var dlg *walk.Dialog

	xs := []Widget{
		Label{Text: "COM порт"},
		Label{Text: "Имя"},
		Label{Text: "Таймаут ответа, мс"},
		Label{Text: "Таймаут байта, мс"},
		Label{Text: "Макс.повторов"},
	}
	tmp := x.config
	availablePorts := comport.AvailablePorts()

	for _, x := range []struct {
		p    *comport.Config
		what string
	}{
		{&tmp.SerialPorts.PortProducts, "Приборы"},
		{&tmp.SerialPorts.PortTemperature, "Термокамера"},
		{&tmp.SerialPorts.PortGas, "Пневмоблок"},
	} {
		x := x
		var edReadTimeout,
			edReadByteTimeout,
			edAttempts *walk.NumberEdit
		var cbPort *walk.ComboBox
		xs = append(xs, []Widget{
			Label{
				Text: x.what,
			},
			ComboBox{
				AssignTo: &cbPort,
				Model:    availablePorts,
				Value:    x.p.Port.Name,
				OnCurrentIndexChanged: func() {
					x.p.Port.Name = cbPort.Text()
				},
			},
			NumberEdit{
				AssignTo: &edReadTimeout,
				Value:    x.p.Mode.ReadTimeout.Seconds() * 1000,
				MinValue: 1,
				MaxValue: 100000,
				OnValueChanged: func() {
					x.p.Mode.ReadTimeout = time.Millisecond * time.Duration(edReadTimeout.Value())
				},
			},
			NumberEdit{
				AssignTo: &edReadByteTimeout,
				Value:    x.p.Mode.ReadByteTimeout.Seconds() * 1000,
				MinValue: 1,
				MaxValue: 100000,
				OnValueChanged: func() {
					x.p.Mode.ReadByteTimeout = time.Millisecond * time.Duration(edReadByteTimeout.Value())
				},
			},
			NumberEdit{
				AssignTo: &edAttempts,
				Value:    float64(x.p.Mode.MaxAttemptsRead),
				MinValue: 1,
				MaxValue: 100,
				OnValueChanged: func() {
					x.p.Mode.MaxAttemptsRead = int(edAttempts.Value())
				},
			},
		}...)

	}

	check(Dialog{
		Icon:          NewIconFromResourceId(IconSettingsID),
		AssignTo:      &dlg,
		Title:         "Настройки",
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		MinSize:       Size{300, 300},
		Layout:        HBox{},
		Font:          Font{PointSize: 10},
		FixedSize:     true,

		Children: []Widget{
			ScrollView{
				HorizontalFixed: true,
				Layout: Grid{
					Columns: 5,
				},
				Children: xs,
			},
			ScrollView{
				HorizontalFixed: true,
				Layout:          VBox{},
				Children: []Widget{
					PushButton{
						Text:     "Применить",
						AssignTo: &acceptPB,
						OnClicked: func() {
							x.config = tmp
							x.SaveConfig()
							dlg.Accept()
						},
					},
					PushButton{
						Text:     "Отмена",
						AssignTo: &cancelPB,
						OnClicked: func() {
							dlg.Cancel()
						},
					},
				},
			},
		},
	}.Create(x.mw))

	dlg.Run()

}
