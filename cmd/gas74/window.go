package main

import (
	"context"
	"fmt"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/comm/modbus"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/powerman/structlog"
	"sync"
	"time"
)

func runMainWindow() {

	settings := walk.NewIniFileSettings("settings.ini")
	defer log.ErrIfFail(settings.Save)

	app := walk.App()
	app.SetOrganizationName("analitpribor")
	app.SetProductName("gas73")
	app.SetSettings(settings)

	log.ErrIfFail(settings.Load)

	if _, err := (MainWindow{
		AssignTo:   &mainWindow,
		Layout:     VBox{},
		Font:       Font{PointSize: 12, Family: "Segoe UI"},
		Background: SolidColorBrush{Color: walk.RGB(0xFF, 0xFF, 0xFF)},
		Size: Size{
			Width:  300,
			Height: 300,
		},
		Children: []Widget{
			Composite{
				//VerticalFixed:true,
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					Label{Text: "СОМ порт:"},
					newComboBoxComport(nil, "COMPORT"),
					PushButton{
						Text: "Установить",
						OnClicked: func() {
							setStatusOk("переключение...")
							go func() {
								if err := setupValves(); err != nil {
									setStatusError(err)
								} else {
									setStatusOk("успешно")
								}
							}()
						},
					},
				},
			},
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					GroupBox{
						Title:  "Вход 1",
						Layout: VBox{},
						Children: []Widget{
							RadioButtonGroup{
								Buttons: []RadioButton{
									newBtn(0, 0),
									newBtn(0, 1),
									newBtn(0, 2),
									newBtn(0, 3),
								},
							},
						},
					},
					GroupBox{
						Layout: VBox{},
						Title:  "Вход 2",
						Children: []Widget{
							RadioButtonGroup{
								Buttons: []RadioButton{
									newBtn(1, 0),
									newBtn(1, 1),
									newBtn(1, 2),
									newBtn(1, 3),
								},
							},
						},
					},
				},
			},

			TextEdit{
				AssignTo: &lblStatus,
			},
		},
	}).Run(); err != nil {
		panic(err)
	}
}

var (
	mainWindow *walk.MainWindow
	lblStatus  *walk.TextEdit
	btnSwitch  [2][4]*walk.RadioButton
	muComport  sync.Mutex
	valveIn    [2]int

	ctx     = context.Background()
	log     = structlog.New()
	comPort = comport.NewPort(func() comport.Config {
		portName, _ := walk.App().Settings().Get("COMPORT")
		return comport.Config{
			Name:        portName,
			Baud:        9600,
			ReadTimeout: time.Millisecond,
		}
	})
)

func portReader() modbus.ResponseReader {
	return comPort.NewResponseReader(ctx, comm.Config{
		ReadByteTimeoutMillis: 50,
		ReadTimeoutMillis:     2000,
		MaxAttemptsRead:       1,
	})
}

func newBtn(in, out int) RadioButton {
	what1 := fmt.Sprintf("Выход %d", out)
	if out == 0 {
		what1 = "Выкл."
	}
	what := fmt.Sprintf("Вход %d - %s", in+1, what1)
	if out == 0 {
		what += "   "
	}
	return RadioButton{
		AssignTo: &btnSwitch[in][out],
		Text:     what1,
		OnClicked: func() {
			valveIn[in] = out
			in2 := 0
			if in == 0 {
				in2 = 1
			}
			for i := 1; i < 4; i++ {
				btnSwitch[in2][i].SetEnabled(out == 0 || out != i)
			}
			btnSwitch[in][out].SetChecked(true)
			if out != 0 {
				btnSwitch[in2][out].SetChecked(false)
			}
		},
	}
}

func setStatus(ok bool, s string) {
	mainWindow.Synchronize(func() {
		c := walk.RGB(0, 0, 0)
		if !ok {
			c = walk.RGB(255, 0, 0)
		}
		lblStatus.SetTextColor(c)
		_ = lblStatus.SetText(s)
	})
}

func setStatusOk(s string) {
	setStatus(true, s)
}
func setStatusError(err error) {
	setStatus(false, err.Error())
}

func setupValves() error {
	c := [4]byte{0x00, 0x11, 0x22, 0x44}
	req := modbus.Request{
		Addr:     0x10,
		ProtoCmd: 0x10,
		Data: []byte{
			0x00, 0x32, 0x00, 0x01, 0x02, c[valveIn[0]], c[valveIn[1]],
		},
	}
	_, err := req.GetResponse(log, portReader(), nil)
	return err
}

func newComboBoxComport(comboBox **walk.ComboBox, key string) ComboBox {
	if comboBox == nil {
		var cb *walk.ComboBox
		comboBox = &cb
	}
	return ComboBox{
		AssignTo:     comboBox,
		Model:        getComports(),
		CurrentIndex: comportIndex(getIniValue(key)),
		OnMouseDown: func(_, _ int, _ walk.MouseButton) {
			cb := *comboBox
			n := cb.CurrentIndex()
			_ = cb.SetModel(getComports())
			_ = cb.SetCurrentIndex(n)
		},
		OnCurrentIndexChanged: func() {
			_ = walk.App().Settings().Put(key, (*comboBox).Text())
		},
	}
}

func getIniValue(key string) string {
	s, _ := walk.App().Settings().Get(key)
	return s
}

func getComports() []string {
	ports, _ := comport.Ports()
	return ports
}

func comportIndex(portName string) int {
	ports, _ := comport.Ports()
	for i, s := range ports {
		if s == portName {
			return i
		}
	}
	return -1
}
