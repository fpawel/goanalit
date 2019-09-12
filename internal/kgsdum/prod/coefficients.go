package main

import (
	"fmt"
	coef2 "github.com/fpawel/goanalit/internal/coef"
	products2 "github.com/fpawel/goanalit/internal/kgsdum/products"
	wask62 "github.com/fpawel/goanalit/internal/kgsdum/wask6"
	"os/exec"
	"path/filepath"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"

	"github.com/fpawel/gutils/utils"
	"github.com/fpawel/gutils/walkUtils"

	"log"
	"strconv"
)

func (x *App) ExecuteEnterCoefficientsRangeDialog() (string, int) {
	var dlg *walk.Dialog
	var ed *walk.LineEdit
	var btnOkCount *walk.PushButton
	resultStr := "0-50"

	dlgResult, err := Dialog{
		AssignTo: &dlg,
		Title:    "Считывание коэффициентов",
		Layout:   VBox{},
		Font:     Font{PointSize: 14},
		Children: []Widget{

			ScrollView{
				VerticalFixed: true,
				Layout:        HBox{SpacingZero: true, MarginsZero: true},
				Children: []Widget{
					Label{Text: "Считывание коэффициентов."},
				},
			},
			ScrollView{
				VerticalFixed: true,
				Layout:        HBox{SpacingZero: true, MarginsZero: true},
				Children: []Widget{
					Label{Text: "Ведите диапазон номеров."},
				},
			},

			LineEdit{
				AssignTo: &ed,
				Text:     x.config.CoefsStr,
				OnTextChanged: func() {
					resultStr = ed.Text()
				},
				OnKeyDown: func(key walk.Key) {
					switch key {
					case walk.KeyReturn:
						dlg.Accept()
					case walk.KeyEscape:
						dlg.Cancel()
					}
				},
			},

			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{
						Text:     "Ок",
						AssignTo: &btnOkCount,
						Image:    AssetImage("assets/png16/checkmark.png"),
						OnClicked: func() {
							dlg.Accept()
						},
					},
					PushButton{
						Text:  "Отмена",
						Image: AssetImage("assets/png16/cancel16.png"),
						OnClicked: func() {
							dlg.Cancel()
						},
					},
				},
			},
		},
	}.Run(x.mw)
	check(err)
	if dlgResult == win.IDOK {
		x.config.CoefsStr = resultStr
	}
	return resultStr, dlgResult

}

func (x *App) ExecuteOpenCoefficientsDialog() coef2.AddrCoefficientValues {
	dlg := &walk.FileDialog{
		FilePath: filepath.Join(x.config.SaveReportsDir, "коэффициенты_КГС_ДУМ.xlsx"),
		Title:    "Открыть к-ты",
		Filter:   "Таблица Excel|*.xlsx",
	}
	ok, err := dlg.ShowOpen(x.mw)
	check(err)
	if !ok {
		return nil
	}
	x.config.SaveReportsDir = filepath.Dir(dlg.FilePath)

	values, err := coef2.OpenFromFile(dlg.FilePath)
	if err != nil {
		walk.MsgBox(x.mw, "КГС ДУМ", err.Error(), walk.MsgBoxIconError|walk.MsgBoxOK)
		return nil
	}
	return values
}

func (x *App) ExecuteReadCoefficientsDialog() coef2.AddrCoefficientValues {
	str, dlgr := x.ExecuteEnterCoefficientsRangeDialog()
	if dlgr != win.IDOK {
		return nil
	}
	xs := utils.ParseIntRanges(str)

	x.config.CoefsStr = utils.FormatIntRanges(xs)

	var coeficients []coef2.Coefficient
	for _, v := range xs {
		coeficients = append(coeficients, coef2.Coefficient(v))
	}

	values := make(coef2.AddrCoefficientValues)

	for row, p := range x.db.Products() {
		if !x.tableProductsModel.Checked(row) {
			continue
		}
		for _, k := range coeficients {
			values[coef2.AddrCoefficient{Addr: p.Addr, Coefficient: k}] = 0
		}
	}
	return values
}

func (x *App) WriteCoefficients() {
	x.ProcessCoefficients(wask62.IODirWrite)
}

func (x *App) ReadCoefficients() {
	x.ProcessCoefficients(wask62.IODirRead)
}

func (x *App) ProcessCoefficients(dir wask62.IODir) {

	var values coef2.AddrCoefficientValues
	if dir == wask62.IODirWrite {
		values = x.ExecuteOpenCoefficientsDialog()
	} else {
		values = x.ExecuteReadCoefficientsDialog()
	}
	if values == nil {
		return
	}

	var what string
	{
		var ps, cs []int
		for _, v := range values.Addresses() {
			ps = append(ps, int(v))
		}
		for _, v := range values.Coefficients() {
			cs = append(cs, int(v))
		}
		what = fmt.Sprintf("%v коэффициентов %s %s", dir, utils.FormatIntRanges(ps), utils.FormatIntRanges(cs))
	}

	x.mw.progressBar.SetVisible(true)
	x.mw.progressBar.SetValue(0)
	x.mw.progressBar.SetRange(0, len(values))
	x.RunComportsTask(NewSimpleProductsTaskConfig(what), func() error {
		defer x.mw.Synchronize(func() {
			x.mw.progressBar.SetVisible(false)
		})

		var hasErrors bool
		for pc, value := range values {
			value := value
			pc := pc
			x.mw.tblProducts.Synchronize(func() {
				x.tableProductsModel.SetSurveyColRow(-1, x.db.Products().ProductRowByAddress(pc.Addr))
			})
			var err error
			switch dir {
			case wask62.IODirRead:
				value, err = wask62.ReadCoefficient(pc, x.PortProducts())

			case wask62.IODirWrite:
				err = wask62.WriteCoefficient(value, pc, x.PortProducts())

			default:
				log.Fatalln("wrong io dir code", dir)
			}

			text, level := walkUtils.ErrorLevel(err, strconv.FormatFloat(float64(value), 'f', -1, 32))
			var tmp products2.ProductTime
			x.WriteTestProductLog(tmp, pc.Addr, products2.TestLogRecord{
				Text:  text,
				Level: level,
				Test:  fmt.Sprintf("%v коэффициента %d", dir, pc.Coefficient),
			})

			progress := 1
			if err != nil {
				hasErrors = true
				if err == fetch.ErrorNoAnswer {
					for pp := range values {
						if pc.Addr == pp.Addr {
							delete(values, pp)
							progress += 1
						}
					}
				}
				p, ok := x.db.Products().ProductByAddress(pc.Addr)
				if ok {
					x.tableProductsModel.AddProductError(p.ProductTime, err)
				}
			} else {
				values[pc] = value
			}

			x.mw.Synchronize(func() {
				connInfo := walkUtils.MessageFromError(err, "успешно")
				connInfo.Text = fmt.Sprintf("%v: %s", pc, strconv.FormatFloat(float64(value), 'f', -1, 32))
				x.mw.lblWorkMessage.SetText(fmt.Sprintf("%v: %s", pc, strconv.FormatFloat(float64(value), 'f', -1, 32)))
				x.mw.progressBar.SetValue(x.mw.progressBar.Value() + progress)

			})

			if err == fetch.ErrorCanceled {
				break
			}
		}

		x.mw.Synchronize(func() {
			if dir == wask62.IODirRead && len(values) > 0 {
				dlg := &walk.FileDialog{
					FilePath: filepath.Join(x.config.SaveReportsDir, "коэффициенты_КГС_ДУМ.xlsx"),
					Title:    "Сохранить к-ты",
					Filter:   "Таблица Excel|*.xlsx",
				}
				ok, err := dlg.ShowSave(x.mw)
				check(err)
				if !ok {
					return
				}
				x.config.SaveReportsDir = filepath.Dir(dlg.FilePath)
				if err := values.SaveToFile(dlg.FilePath); err != nil {
					walk.MsgBox(x.mw, what, err.Error(), walk.MsgBoxIconError|walk.MsgBoxOK)
					return
				}
				cmd := exec.Command("explorer.exe", "/e,", "/select,", dlg.FilePath)
				err = cmd.Start()
				if err != nil {
					walk.MsgBox(x.mw, what, err.Error(), walk.MsgBoxIconError|walk.MsgBoxOK)
				} else {
					cmd.Wait()
				}
			}
		})

		if hasErrors {
			return fmt.Errorf("произошла одна или несколько ошибок")
		}
		return nil
	})

}
