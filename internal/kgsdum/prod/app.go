package main

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	products2 "github.com/fpawel/goanalit/internal/kgsdum/products"
	productsView2 "github.com/fpawel/goanalit/internal/kgsdum/productsView"
	"github.com/fpawel/guartutils/comport"
	"github.com/lxn/walk"
	"github.com/lxn/win"
	"io/ioutil"
	"os"
	"path/filepath"
)

type App struct {
	*walk.Application
	db                  products2.DB
	mw                  *AppMainWindow
	cancellationDelay   int32
	tableProductsModel  *TableProductsModel
	tableMainWorksModel *TableMainWorksModel
	tableLogsModel      *productsView2.TableLogsViewModel
	config              AppConfig
	ports               *comport.Ports
}

func NewApp() *App {
	x := &App{
		Application: walk.App(),
		config:      NewAppConfig(),
		ports:       comport.NewPorts(),
	}
	x.mw = NewAppMainWindow(x)

	x.SetOrganizationName("Аналитприбор")
	x.SetProductName("КГС-ДУМ")
	fmt.Println(x.FolderPath())

	{
		sets := walk.NewIniFileSettings("settings.ini")
		if err := sets.Load(); err != nil {
			println("load settings.ini error:", err)
		}
		x.SetSettings(sets)
	}

	{
		settings := walk.NewIniFileSettings("settings.ini")
		if err := settings.Load(); err != nil {
			fmt.Println("load settings.ini error:", err)
		}
		x.SetSettings(settings)
	}

	// считать настройки приложения из сохранённого файла json
	{
		b, err := ioutil.ReadFile(x.ConfigPath())
		if err != nil {
			fmt.Print("config.json error:", err)
		} else {
			if err := json.Unmarshal(b, &x.config); err != nil {
				fmt.Print("config.json content error:", err)
			}
		}
	}

	// создать каталог с данными и настройками программы если его нет
	if _, err := os.Stat(x.FolderPath()); os.IsNotExist(err) {
		os.Mkdir(x.FolderPath(), os.ModePerm)
	}

	var err error
	x.db.DB, err = bolt.Open(x.FolderPath()+"/products.db", 0600, nil)
	check(err)

	x.tableProductsModel = NewTableProductsModel(x)
	x.tableMainWorksModel = NewTableWorksViewModel(x)

	var partyTime products2.PartyTime
	x.db.View(func(tx products2.Tx) {
		partyTime = tx.Party().PartyTime
	})

	x.tableLogsModel = productsView2.NewTableLogsViewModel(x.db, partyTime)
	check(x.mw.Markup().Create())
	x.mw.Initialize()

	return x
}

func (x *App) ProductValue(p products2.Product, row int, col int) (r interface{}) {
	switch col {
	case 1:
		r = p.Addr()
	case 2:
		r = p.Serial()
	}
	return
}

func (x *App) FolderPath() string {

	appDataPath, err := walk.AppDataPath()
	check(err)
	return filepath.Join(
		appDataPath,
		x.OrganizationName(),
		x.ProductName())
}

func (x *App) ConfigPath() string {
	return filepath.Join(x.FolderPath(), "config.json")
}

func (x *App) Close() {
	check(x.Settings().(*walk.IniFileSettings).Save())
	check(x.db.DB.Close())
	x.SaveConfig()
}

func (x *App) SaveConfig() {
	x.config.Save(x.ConfigPath())
}

func (x *App) UpdateTableLogs() {
	x.mw.Synchronize(func() {
		x.tableLogsModel.PublishRowsReset()
		x.mw.tblLogs.SendMessage(win.WM_VSCROLL, win.SB_PAGEDOWN, 0)
	})
}
