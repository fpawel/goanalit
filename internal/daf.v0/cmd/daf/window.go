package main

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/daf.v0/internal/data"
	"github.com/fpawel/daf.v0/internal/viewmodel"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	"github.com/powerman/structlog"
	"math"
	"time"
)

type DafMainWindow struct {
	*walk.MainWindow
	lblWork,
	lblWorkTime *walk.Label
	DelayHelp *delayHelp
}

func (mw DafMainWindow) SetWorkStatus(color walk.Color, text string) {
	mw.Synchronize(
		func() {
			_ = mw.lblWorkTime.SetText(time.Now().Format("15:04:05"))
			_ = mw.lblWork.SetText(text)
			mw.lblWork.SetTextColor(color)
		})

}

func (mw DafMainWindow) IgnoreErrorPrompt(title string, err error) (result bool) {
	if merry.Is(err, context.Canceled) {
		return false
	}
	ch := make(chan struct{})
	mw.Synchronize(func() {
		result = walk.MsgBox(dafMainWindow.MainWindow, title,
			err.Error()+"\n\nИгнорировать ошибку и продолжить?",
			walk.MsgBoxIconError|walk.MsgBoxYesNo) == win.IDYES
		ch <- struct{}{}
	})
	<-ch
	return
}

func runMainWindow() error {

	app := walk.App()
	app.SetOrganizationName("analitpribor")
	app.SetProductName("EN8800-6408")
	settings := walk.NewIniFileSettings("settings.ini")
	if err := settings.Load(); err != nil {
		panic(err)
	}
	app.SetSettings(settings)

	getIniValue := func(key string) string {
		s, _ := settings.Get(key)
		return s
	}

	newComboBoxComport := func(comboBox **walk.ComboBox, key string) ComboBox {
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
				_ = settings.Put(key, (*comboBox).Text())
			},
		}
	}

	var (
		cbComportDaf,
		cbComportHart *walk.ComboBox
		tblViewProducts,
		tblViewProductValues,
		tblViewProductEntries *walk.TableView
		gbProductValues,
		gbProductEntries *walk.GroupBox
		neCmd, neArg *walk.NumberEdit
		pbCancelWork *walk.PushButton
		btnRun       *walk.SplitButton
		gbCmd        *walk.GroupBox
	)

	prodsMdl = viewmodel.NewDafProductsTable(func(f func()) {
		tblViewProducts.Synchronize(f)
	})

	showErr := func(title, text string) {
		walk.MsgBox(dafMainWindow.MainWindow, title,
			text, walk.MsgBoxIconError|walk.MsgBoxOK)
	}

	dafMainWindow.DelayHelp = new(delayHelp)

	var workStarted bool
	doWork := func(what string, work func() error) {
		if workStarted {
			panic("already started")
		}
		workStarted = true

		prodsMdl.ClearConnectionsInfo()

		ctxApp, cancelWorkFunc = context.WithCancel(context.Background())
		btnRun.SetVisible(false)
		gbCmd.SetVisible(false)

		pbCancelWork.SetVisible(true)
		dafMainWindow.SetWorkStatus(ColorNavy, what+": выполняется")

		go func() {
			err := work()

			_ = portHart.Close()
			_ = portDaf.Close()
			dafMainWindow.MainWindow.Synchronize(func() {
				workStarted = false

				gbCmd.SetVisible(true)
				btnRun.SetVisible(true)

				pbCancelWork.SetVisible(false)
				prodsMdl.SetInterrogatePlace(-1)

				if err != nil {
					if merry.Is(err, context.Canceled) {
						dafMainWindow.SetWorkStatus(walk.RGB(139, 69, 19), what+": прервано")
					} else {
						dafMainWindow.SetWorkStatus(walk.RGB(255, 0, 0), what+": "+err.Error())

						structlog.New().PrintErr(err)

						showErr(what, err.Error())
					}

				} else {
					dafMainWindow.SetWorkStatus(ColorNavy, what+": выполнено")
				}
			})
		}()
	}

	actionWork := func(what string, f func() error) Action {
		return Action{
			Text: what,
			OnTriggered: func() {
				doWork(what, f)
			},
		}
	}

	prodValuesMdl := viewmodel.NewDafProductValuesTable()
	prodEntriesMdl := viewmodel.NewDafProductEntriesTable()

	if err := (MainWindow{
		AssignTo: &dafMainWindow.MainWindow,
		Title: "ЭН8800-6408 Партия ДАФ-М " + (func() string {
			p := data.GetLastParty()
			return fmt.Sprintf("№%d %s", p.PartyID, p.CreatedAt.Format("02.01.2006"))
		}()),
		Name:       "MainWindow",
		Font:       Font{PointSize: 12, Family: "Segoe UI"},
		Background: SolidColorBrush{Color: walk.RGB(255, 255, 255)},
		Size:       Size{800, 600},
		Layout:     VBox{},

		MenuItems: []MenuItem{
			Menu{
				Text: "Окно",
				Items: []MenuItem{
					Action{
						Text: "Консоль",
						OnTriggered: func() {
							formConsoleShow()
							//formConsoleNewLine(DBG, "дебаг")
							//formConsoleNewLine(INF, "инфа")
							//formConsoleNewLine(ERR, "ошибочка")
							//formConsoleNewLine(WRN, "варнинг")
						},
					},
				},
			},
		},

		Children: []Widget{
			ScrollView{
				VerticalFixed: true,
				Layout:        HBox{},
				Children: []Widget{
					SplitButton{
						Text: "Партия",
						MenuItems: []MenuItem{
							Action{
								Text: "Создать новую",
								OnTriggered: func() {
									if walk.MsgBox(dafMainWindow.MainWindow, "Новая партия",
										"Подтвердите необходимость создания новой партии",
										walk.MsgBoxIconQuestion|walk.MsgBoxYesNo) != win.IDYES {
										return
									}

									data.CreateNewParty()
									prodsMdl.Validate()
								},
							},
							Action{
								Text: "Параметры",
								OnTriggered: func() {
									runPartyDialog(dafMainWindow.MainWindow)
								},
							},
							Action{
								Text: "Добавить прибор в партию",
								Shortcut: Shortcut{
									Key: walk.KeyInsert,
								},
								OnTriggered: func() {
									prodsMdl.AddNewProduct()
								},
							},
						},
					},
					SplitButton{
						Text:     "Управление",
						AssignTo: &btnRun,
						MenuItems: []MenuItem{
							actionWork("Опрос", interrogateProducts),
							actionWork("Настройка ДАФ-М", dafSetupMain),
							Separator{},
							actionWork("Настройка токового выхода", dafSetupCurrent),
							actionWork("Установка порогов для настройки", dafSetupThresholdTest),
							actionWork("Корректировка показаний", dafAdjust),
							actionWork("Проверка диапазона измерений", dafTestMeasureRange),
							actionWork("Проверка HART протокола", dafTestMeasureRange),
							Separator{},
							actionWork("Подать ПГС1", func() error { return switchGas(1) }),
							actionWork("Подать ПГС2", func() error { return switchGas(2) }),
							actionWork("Подать ПГС3", func() error { return switchGas(3) }),
							actionWork("Подать ПГС4", func() error { return switchGas(4) }),
							actionWork("Отключить газ", func() error { return switchGas(0) }),
						},
					},
					PushButton{
						AssignTo: &pbCancelWork,
						Text:     "Прервать",
						OnClicked: func() {
							cancelWorkFunc()
						},
					},

					Label{
						AssignTo:  &dafMainWindow.lblWorkTime,
						TextColor: walk.RGB(0, 128, 0),
					},
					Label{
						AssignTo: &dafMainWindow.lblWork,
					},
					dafMainWindow.DelayHelp.Widget(),
				},
			},
			ScrollView{
				Layout: HBox{MarginsZero: true, SpacingZero: true},
				Children: []Widget{

					GroupBox{
						Layout: Grid{},
						Title:  "Настраиваемые газоанализаторы ДАФ-М",
						Children: []Widget{
							TableView{
								AssignTo:                 &tblViewProducts,
								NotSortableByHeaderClick: true,
								LastColumnStretched:      true,
								CheckBoxes:               true,
								Model:                    prodsMdl,
								OnItemActivated: func() {
									n := tblViewProducts.CurrentIndex()
									if n < 0 || n >= prodsMdl.RowCount() {
										return
									}
									runProductDialog(dafMainWindow.MainWindow, prodsMdl.ProductAt(n))
									prodsMdl.PublishRowChanged(n)
								},
								OnKeyDown: func(key walk.Key) {
									switch key {

									case walk.KeyInsert:
										m := prodsMdl
										m.AddNewProduct()
										runProductDialog(dafMainWindow.MainWindow, m.ProductAt(m.RowCount()-1))
										prodsMdl.PublishRowChanged(m.RowCount() - 1)

									case walk.KeyDelete:
										n := tblViewProducts.CurrentIndex()
										m := prodsMdl
										if n < 0 || n >= m.RowCount() {
											return
										}
										if err := data.DBProducts.Delete(m.ProductAt(n)); err != nil {
											showErr("Ошибка данных", err.Error())
										}
										prodsMdl.Validate()
									}

								},
								OnCurrentIndexChanged: func() {
									n := tblViewProducts.CurrentIndex()
									if n > -1 && n < prodsMdl.RowCount() {
										p := prodsMdl.ProductAt(n)
										s := fmt.Sprintf("ДАФ-М № %d, заводской номер %d, адрес %d", p.ProductID, p.Serial, p.Addr)
										_ = gbProductValues.SetTitle("Паспорт " + s)
										_ = gbProductEntries.SetTitle("Журнал " + s)

										prodValuesMdl.SetProduct(prodsMdl.ProductAt(n).ProductID)
										prodEntriesMdl.SetProduct(prodsMdl.ProductAt(n).ProductID)

										gbProductValues.SetVisible(prodValuesMdl.RowCount() > 0)
										gbProductEntries.SetVisible(prodEntriesMdl.RowCount() > 0)

										return
									}

									gbProductValues.SetVisible(false)
									gbProductEntries.SetVisible(false)
								},

								Columns: viewmodel.ProductColumns,
							},
						},
					},
					ScrollView{
						HorizontalFixed: true,
						Layout:          VBox{},
						Children: []Widget{
							GroupBox{
								Title:  "COM порты",
								Layout: VBox{},
								Children: []Widget{
									Label{Text: "Стенд и приборы:"},
									newComboBoxComport(&cbComportDaf, "COMPORT_PRODUCTS"),
									Label{Text: "HART модем:"},
									newComboBoxComport(&cbComportHart, "COMPORT_HART"),
								},
							},
							GroupBox{
								AssignTo: &gbCmd,
								Layout:   VBox{},
								Title:    "Команда:",
								Children: []Widget{
									Label{Text: "Код:"},
									NumberEdit{
										AssignTo: &neCmd,
										MinValue: 1,
										MaxValue: math.MaxFloat64,
									},
									Label{Text: "Аргумент:"},
									NumberEdit{AssignTo: &neArg, Decimals: 2, MinSize: Size{80, 0}},
									PushButton{Text: "Выполнить", OnClicked: func() {
										cmd := modbus.DevCmd(neCmd.Value())
										arg := neArg.Value()
										doWork(fmt.Sprintf("Оправка команды %d,%v", cmd, arg), func() error {
											return dafSendCmdToEachOkProduct(cmd, arg)
										})
									}},
								},
							},
						},
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{

					GroupBox{
						AssignTo: &gbProductValues,
						Layout:   Grid{},
						Title:    "Паспорт",
						Children: []Widget{
							TableView{
								AssignTo:                 &tblViewProductValues,
								NotSortableByHeaderClick: true,
								LastColumnStretched:      true,
								Model:                    prodValuesMdl,
								Columns:                  viewmodel.ProductValueColumns,
							},
						},
					},

					GroupBox{
						AssignTo: &gbProductEntries,
						Layout:   Grid{},
						Title:    "Журнал",
						Children: []Widget{
							TableView{
								AssignTo:                 &tblViewProductEntries,
								NotSortableByHeaderClick: true,
								LastColumnStretched:      true,
								Model:                    prodEntriesMdl,
								Columns:                  viewmodel.ProductEntryColumns,
							},
						},
					},
				},
			},
		},
	}).Create(); err != nil {
		return err
	}

	pbCancelWork.SetVisible(false)
	prodsMdl.Validate()

	dafMainWindow.MainWindow.Run()

	if err := settings.Save(); err != nil {
		return err
	}
	return nil
}

func runProductDialog(owner walk.Form, p *data.Product) {
	var (
		edAddr, edSerial *walk.NumberEdit
		dlg              *walk.Dialog
		btn              *walk.PushButton
		lblError         *walk.Label
		saveOnEdit       = false
	)

	save := func(what string) {
		if !saveOnEdit {
			return
		}
		p.Serial = int64(edSerial.Value())
		p.Addr = modbus.Addr(edAddr.Value())
		_ = lblError.SetText("")
		if err := data.DBProducts.Save(p); err != nil {
			_ = lblError.SetText(fmt.Sprintf("%s: дублирование значения: %v", what, err))
			if err := data.DBProducts.FindByPrimaryKeyTo(p, p.ProductID); err != nil {
				panic(err)
			}
		}
		if edSerial.Value() != float64(p.Serial) {
			edSerial.SetTextColor(0xFF)
		} else {
			edSerial.SetTextColor(0)
		}

		if edAddr.Value() != float64(p.Addr) {
			edAddr.SetTextColor(0xFF)
		} else {
			edAddr.SetTextColor(0)
		}

	}
	d := Dialog{
		Title:         fmt.Sprintf("ДАФ %d", p.ProductID),
		Font:          Font{PointSize: 12, Family: "Segoe UI"},
		Background:    SolidColorBrush{Color: walk.RGB(255, 255, 255)},
		Layout:        Grid{Columns: 2},
		AssignTo:      &dlg,
		DefaultButton: &btn,
		CancelButton:  &btn,
		Children: []Widget{
			Label{Text: "Адрес:", TextAlignment: AlignFar},
			NumberEdit{
				AssignTo: &edAddr,
				Value:    float64(p.Addr),
				MinValue: 1,
				MaxValue: 127,
				Decimals: 0,
				OnValueChanged: func() {
					save(fmt.Sprintf("адрес: %v", edAddr.Value()))
				},
			},
			Label{Text: "Серийный номер:", TextAlignment: AlignFar},
			NumberEdit{
				AssignTo: &edSerial,
				Value:    float64(p.Serial),
				MinValue: 1,
				MaxValue: math.MaxFloat64,
				Decimals: 0,
				OnValueChanged: func() {
					save(fmt.Sprintf("cерийный номер: %v", edSerial.Value()))
				},
			},
			Composite{
				Layout: HBox{},
			},
			PushButton{
				AssignTo: &btn,
				Text:     "Закрыть",
				OnClicked: func() {
					dlg.Accept()
				},
			},
			Label{
				ColumnSpan: 2,
				AssignTo:   &lblError,
				TextColor:  0xFF,
			},
		},
	}
	if err := d.Create(owner); err != nil {
		walk.MsgBox(owner, fmt.Sprintf("ДАФ %d", p.ProductID), err.Error(), walk.MsgBoxIconError|walk.MsgBoxOK)
		return
	}
	saveOnEdit = true
	dlg.Run()
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

var ColorNavy = walk.RGB(0, 0, 128)
