package internal

import (
	"github.com/lxn/win"
	"github.com/tarm/serial"
	"time"
)

type Worker struct {
	m          *Model
	portConfig *serial.Config
	port       *serial.Port
	quit       chan struct{}
	chart      *PipeChartServer
}

func NewWorker(m *Model) (x *Worker) {
	x = &Worker{
		portConfig: &serial.Config{Baud: 115200, ReadTimeout: 500 * time.Millisecond},
		port:       &serial.Port{},
		m:          m,
		quit:       make(chan struct{}, 1),
		chart:      NewPipeChartServer(),
	}
	go x.chart.Run(func() {
		x.quit <- struct{}{}
		x.m.config.mainWindow.SafeClose()
	})

	x.m.RestorePartySeries(x.m.party.chart, x.chart) // загрузить сохранённые точки в чарт
	x.chart.Send(2, []byte{0})
	return
}

func (x *Worker) setupPort() {
	mw := x.m.config.mainWindow
	portName := mw.getSerialPortName()

	if x.portConfig.Name != portName {
		if x.port != nil {
			x.port.Close()
			x.port = nil
		}
	}

	if portName == "" {
		mw.setNotifyMessage(&NotifyMessage{
			"СОМ порт не задан. Вам следует указать имя СОМ порта, к которому подключен стенд.",
			win.NIIF_WARNING,
		})
		return
	}

	if x.port == nil {
		x.portConfig.Name = portName
		var err error
		x.port, err = serial.OpenPort(x.portConfig)
		if err != nil {
			mw.setComportError(err)
			x.port = nil
		}
	}
}

func (x *Worker) Stop() {
	x.quit <- struct{}{}
}

func (x *Worker) Run() {

	mw := x.m.config.mainWindow

	x.port, _ = serial.OpenPort(&serial.Config{Name: "COM1", Baud: 115200, ReadTimeout: 500 * time.Millisecond})

	var buffRxD [53]byte
	var buffTxD [8]byte

	for {
		for n := byte(0); n < 5; n++ {
			select {
			case <-x.quit:
				x.chart.Stop()
				return
			default:

				x.setupPort()
				if x.port == nil {
					continue
				}

				mw.setComportOk()

				errPort, errStend := getStendData(x.port, x.portConfig.Name, byte(n+1), buffTxD[:], buffRxD[:])
				x.port.Flush()

				if errPort != nil {
					mw.setComportError(errPort)
					continue
				}

				var storeValues []TimeValue
				if errStend == nil {
					x.m.party.conn[n] = StendConnceted
					x.m.party.errCount[n] = 0

					if n == 0 {
						temp, tempOk := parseBCD(buffRxD[3+4*10:])
						pres, presOk := parseBCD(buffRxD[3+4*11:])
						mw.setTemperature(temp, tempOk)
						mw.setPressure(pres, presOk)
						if tempOk {
							v := TimeValue{0, time.Now(), temp}
							x.chart.SendTimeValue(PipeChartCmdAddCurrentTimeValue, v)
							storeValues = append(storeValues, v)

						}
						if presOk {
							v := TimeValue{1, time.Now(), pres}
							x.chart.SendTimeValue(PipeChartCmdAddCurrentTimeValue, v)
							storeValues = append(storeValues, v)
						}
					}

					for i := byte(0); i < 10; i++ {
						index := n*10 + i
						p := &x.m.party.products[index]
						offset := 3 + i*4
						bs := buffRxD[offset : offset+4]
						value, ok := parseBCD(bs)
						p.err = !ok

						if ok && value < 60 {
							p.value = &value
							timeValue := TimeValue{index + 2, time.Now(), value}
							x.chart.SendTimeValue(PipeChartCmdAddCurrentTimeValue, timeValue)
							storeValues = append(storeValues, timeValue)
						} else {
							p.value = nil
						}
					}
				} else {
					x.m.party.errCount[n]++
					if x.m.party.errCount[n] > 5 {
						x.m.party.conn[n] = StendConnectionError
					}
				}
				x.m.StoreTimeValues(storeValues)
				mw.refreshStendInfo(int(n))

				mw.tableViewParty.Synchronize(func() {
					x.m.party.PublishRowsReset()
				})

			}
		}

	}
}
