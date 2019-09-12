package main

import (
	kgsdum2 "github.com/fpawel/goanalit/internal/kgsdum/kgsdum"
	products2 "github.com/fpawel/goanalit/internal/kgsdum/products"
	wask62 "github.com/fpawel/goanalit/internal/kgsdum/wask6"
	"github.com/fpawel/gutils/walkUtils"
	"github.com/lxn/walk"
	"strconv"
	"time"

	"fmt"
	"github.com/lxn/win"
)

type RunTaskConfig struct {
	Ports      []ComportType
	What       string
	ShowReport bool
}

func NewSimpleProductsTaskConfig(what string) RunTaskConfig {
	return RunTaskConfig{
		Ports:      []ComportType{PortProducts},
		What:       what,
		ShowReport: true,
	}
}

func (x *App) RunComportsTask(config RunTaskConfig, task func() error) {
	if err := x.OpenPorts(config.Ports); err != nil {
		walk.MsgBox(x.mw, x.ProductName(), config.What+": "+err.Error(), walk.MsgBoxIconWarning)
		return
	}
	x.mw.InvalidateWorkRunning(true)
	x.mw.lblWorkInfo.SetText(config.What)

	go func() {
		x.WriteLogParty(products2.TestLogRecord{
			Test: config.What,
			Text: "начало выполнения",
		})
		err := task()
		x.mw.Synchronize(func() {
			text, level := walkUtils.ErrorLevel(err, "успешно")
			x.WriteLogParty(products2.TestLogRecord{
				Test:  config.What,
				Text:  text,
				Level: level,
			})
			if config.ShowReport {
				stl := walk.MsgBoxIconInformation
				if err != nil {
					stl = walk.MsgBoxIconError
				}
				walk.MsgBox(x.mw, x.ProductName(), config.What+"\n\n"+text, stl)
			}
			x.mw.InvalidateWorkRunning(false)
		})
	}()
}

func (x *App) RunMainWork() {
	workConfig := RunTaskConfig{
		Ports:      []ComportType{PortProducts, PortTemperature, PortGas},
		What:       "Настройка КГС ДУМ",
		ShowReport: true,
	}
	workProvider := workProvider{x}
	x.RunComportsTask(workConfig, func() (err error) {
		defer x.mw.tblWorks.Synchronize(func() {
			x.tableMainWorksModel.SetCurrentWork(nil)
		})
		for i, work := range kgsdum2.Works() {
			if x.ports.Canceled() {
				return fetch.ErrorCanceled
			}
			if !x.tableMainWorksModel.Checked(i) {
				continue
			}

			x.mw.tblWorks.Synchronize(func() {
				x.tableMainWorksModel.SetCurrentWork(work)
				x.mw.lblWorkInfo.SetText(work.String())
			})
			if err = work.Action(workProvider); err != nil {
				x.WriteLogParty(products2.TestLogRecord{
					Test:  work.String(),
					Text:  err.Error(),
					Level: win.NIIF_ERROR,
				})
				break
			}
		}
		return
	})

}

func (x *App) RunSetAddr() {
	addr, ok := ExecuteSetAddrDialog(x.mw)
	if !ok {
		return
	}
	x.RunComportsTask(NewSimpleProductsTaskConfig(fmt.Sprintf("установка адреса %d", addr)), func() (err error) {
		portConfig := x.config.SerialPorts.PortProducts
		_, err = x.ports.Write(portConfig.Port, []byte{0, 0xAA, 0x55, addr})
		if err != nil {
			return
		}
		time.Sleep(time.Second / 2)
		_, err = wask62.ReadVar(wask62.DeviceAddr(addr), wask62.ValueAddr(kgsdum2.VarTemperature), x.PortProducts())
		return err
	})
}

func (x *App) RunSendCommand() {
	cmd, arg, ok := x.ExecuteSendCommandDialog()
	if !ok {
		return
	}
	x.RunComportsTask(NewSimpleProductsTaskConfig(fmt.Sprintf("команда %d, запись %g", cmd, arg)), func() error {
		return x.ProcessEachProduct(fmt.Sprintf("команда %d", cmd), fmt.Sprintf("запись %g", arg), func(p products2.ProductInfo) error {
			return wask62.SendCommand(wask62.DeviceAddr(p.Addr), cmd, arg, x.PortProducts())
		})

	})

}

func (x *App) RunManualSurvey() {
	x.RunComportsTask(RunTaskConfig{
		What:  "опрос приборов",
		Ports: []ComportType{PortProducts},
	}, func() error {
		for {
			var ps []products2.ProductInfo
			for row, p := range x.db.Products() {
				if x.tableProductsModel.Checked(row) {
					ps = append(ps, p)
				}
			}
			if len(ps) == 0 {
				return fmt.Errorf("приборы не выбраны")
			}
			for _, p := range ps {
				p := p
				for nVar, deviceVar := range x.config.Vars() {
					x.mw.tblProducts.Synchronize(func() {
						x.tableProductsModel.SetSurveyColRow(nVar+1, p.Row)
					})
					value, err :=
						wask62.ReadVar(wask62.DeviceAddr(p.Addr), wask62.ValueAddr(deviceVar), x.PortProducts())

					if err == fetch.ErrorCanceled {
						return nil
					}
					x.tableProductsModel.SetProductValue(p.ProductTime, deviceVar, Float32Result{float64(value), err})
					x.mw.Synchronize(func() {
						x.tableProductsModel.PublishRowsReset()
						textColor := walk.RGB(0, 102, 204)
						text := strconv.FormatFloat(float64(value), 'f', -1, 32)
						if err != nil {
							textColor = walk.RGB(255, 0, 0)
							text = err.Error()
						}
						x.mw.lblWorkMessage.SetTextColor(textColor)
						x.mw.lblWorkMessage.SetText(fmt.Sprintf("№%02d, адрес %02d, %v: %s", p.Row+1, p.Addr, deviceVar, text))
					})
				}
			}
		}
		return nil
	})
}
