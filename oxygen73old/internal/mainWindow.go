package internal

import (
	"time"

	"fmt"
	"github.com/lxn/walk"
	"github.com/lxn/win"
)

type AppMainWindow struct {
	*walk.MainWindow
	model               *Model
	chart               *PipeChartServer
	notifyMessage       *NotifyMessageTime
	comportInfo         *NotifyMessageTime
	notifyIcon          *walk.NotifyIcon
	tableViewParty      *walk.TableView
	comboBoxPort        *walk.ComboBox
	newPartyAction      *walk.Action
	browsePartiesAction *walk.Action
	editSerialsAction   *walk.Action
	settingsAction      *walk.Action
	labelPortConn       *walk.Label
	labelPartyCreatedAt *walk.Label
	labelPressure       *walk.Label
	labelTemperature    *walk.Label
	progressOpenChart   *walk.ProgressBar
	panelOpenChart      *walk.Composite
}

type NotifyMessage struct {
	Info  string
	Level uint
}

type NotifyMessageTime struct {
	*NotifyMessage
	Time time.Time
}

func (x *AppMainWindow) initialize() {
	x.refreshSerials()
	// Create the notify icon and make sure we clean it up on exit.
	x.createNotifyIcon()
	x.refreshLabelPartyCreatedAt()

	b, err := walk.NewSolidColorBrush(walk.RGB(204, 229, 255))
	if err != nil {
		logger.Panic(err)
	}
	x.labelTemperature.SetBackground(b)

	b, err = walk.NewSolidColorBrush(walk.RGB(255, 204, 204))
	if err != nil {
		logger.Panic(err)
	}
	x.labelPressure.SetBackground(b)
}

func (x *AppMainWindow) SafeClose() {
	x.Synchronize(func() {
		if err := x.Close(); err != nil {
			logger.Panic(err)
		}
	})
}

func (x *AppMainWindow) refreshLabelPartyCreatedAt() {
	x.labelPartyCreatedAt.SetText(x.model.party.chart.CreatedAt.Format("02 01 2006, 03:04"))
}

func (x *AppMainWindow) refreshStendInfo(n int) {
	x.tableViewParty.Synchronize(func() {
		text := "..."
		if x.model.party.conn[n] == StendConnectionError {
			text = "нет связи"
		} else if x.model.party.conn[n] == StendConnceted {
			text = "+"
		}

		err := x.tableViewParty.Columns().At(n*3 + 2).SetTitle(text)
		if err != nil {
			logger.Panic(err)
		}
	})
	x.labelPortConn.Synchronize(func() {

		str := "связь установлена"
		color := walk.Color(0xF5F5DC)
		for i := range x.model.party.conn {
			if x.model.party.conn[i] == StendNotConnceted {
				str = "связь устанавливается..."
				break
			} else if x.model.party.conn[i] == StendConnectionError {
				str = "нет связи со стендом"
				color = walk.Color(0x00FFFF)
				break
			}
		}

		if x.labelPortConn.Text() != str {
			setLabelColorText(x.labelPortConn, str,
				newFont("Arial", 10, 0),
				color)
		}
	})
}

func (x *AppMainWindow) setComportError(err error) {
	portName := x.getSerialPortName()
	if portName == "" {
		portName = "COM порт"
	}

	x.setNotifyMessage(&NotifyMessage{
		portName + ": " + err.Error(),
		win.NIIF_ERROR})
	x.comportInfo = x.notifyMessage
	if x.labelPortConn.Text() != "нет связи" {

		setLabelColorText(x.labelPortConn, "нет связи",
			newFont("Arial", 10, walk.FontBold),
			walk.Color(0x00FFFF))
	}
}

func (x *AppMainWindow) setComportOk() {

	portName := x.getSerialPortName()
	if portName == "" {
		portName = "COM порт"
	}

	m := &NotifyMessage{
		portName + "открыт",
		win.NIIF_INFO}

	if x.comportInfo != nil && *x.comportInfo.NotifyMessage == *m {
		return
	}

	x.setNotifyMessage(m)
	x.comportInfo = x.notifyMessage

}

func (x *AppMainWindow) setNotifyMessage(m *NotifyMessage) {

	x.Synchronize(func() {
		if x.notifyMessage != nil && *x.notifyMessage.NotifyMessage == *m && time.Since(x.notifyMessage.Time) < 5*time.Second {
			return
		}
		x.notifyMessage = &NotifyMessageTime{
			NotifyMessage: m,
			Time:          time.Now(),
		}

		var f func(title, info string) error
		switch m.Level {
		case win.NIIF_INFO:
			f = x.notifyIcon.ShowInfo
		case win.NIIF_WARNING:
			f = x.notifyIcon.ShowWarning
		case win.NIIF_ERROR:
			f = x.notifyIcon.ShowError
		case win.NIIF_USER:
			f = x.notifyIcon.ShowInfo
		default:
			f = x.notifyIcon.ShowMessage
		}

		if err := f("Лаб.№73. График О2.", m.Info); err != nil {
			logger.Panic(err)
		}
	})

}

func (x *AppMainWindow) createNotifyIcon() {
	// We load our icon from a file.
	icon, err := walk.NewIconFromFile("./img/o2.ico")
	if err != nil {
		logger.Panic(err)
	}

	x.notifyIcon, err = walk.NewNotifyIcon(x.MainWindow)
	if err != nil {
		logger.Panic(err)
	}

	// Set the icon and a tool tip text.
	if err := x.notifyIcon.SetIcon(icon); err != nil {
		logger.Panic(err)
	}
	if err := x.notifyIcon.SetToolTip("Кислородные ЭХЯ, Лаборатория 73"); err != nil {
		logger.Panic(err)
	}

	// The notify icon is hidden initially, so we have to make it visible.
	if err := x.notifyIcon.SetVisible(true); err != nil {
		logger.Panic(err)
	}
}

func (x *AppMainWindow) getSerialPortName() string {
	return x.comboBoxPort.Text()
}

func (x *AppMainWindow) refreshSerials() {
	xs := getSerialPortNames()
	if x.comboBoxPort.Text() != "" {
		xs = append(xs, x.comboBoxPort.Text())
	}
	x.comboBoxPort.SetModel(xs)

}

func (x *AppMainWindow) StyleCell(style *walk.CellStyle) {

	/*
		i := (style.Col()/3)*10 + style.Row()
		p := &x.model.party.products[i]

		if p.err != nil || x.model.party.connErr[i / 10] != nil {
			style.BackgroundColor = walk.RGB(255, 102, 102)
		}
	*/

	switch style.Col() % 3 {
	case 0:
		style.TextColor = walk.RGB(0, 0, 0)
		style.Font = newFont("Arial", 9, walk.FontBold)
	case 1:
		style.TextColor = walk.RGB(128, 128, 128)
		style.Font = newFont("Arial", 8, walk.FontItalic)
	default:

		style.TextColor = walk.RGB(0, 0, 0)
		style.Font = x.tableViewParty.Font()
	}

}

func (x *AppMainWindow) editSerialsAction_Triggered() {
	_, err := x.runSerialsDialog()
	if err != nil {
		logger.Panic(err)
	}
}

func (x *AppMainWindow) aboutAction_Triggered() {

}

func (x *AppMainWindow) setTemperature(v float32, ok bool) {
	x.labelTemperature.Synchronize(func() {
		text := "???"
		if ok {
			text = fmt.Sprintf("%.1f \"C", v)
		}
		x.labelTemperature.SetText(text)
	})
}

func (x *AppMainWindow) setPressure(v float32, ok bool) {
	x.labelPressure.Synchronize(func() {
		text := "???"
		if ok {
			text = fmt.Sprintf("%.1f \"C", v)
		}
		x.labelPressure.SetText(text)
	})
}

func newFont(family string, pointSize int, style walk.FontStyle) *walk.Font {
	f, err := walk.NewFont(family, pointSize, style)
	if err != nil {
		logger.Panic(err)
	}
	return f
}
