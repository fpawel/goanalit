package internal

import (
	"github.com/lxn/walk"
	"os"
	"path/filepath"

	"sync"
)

func Main() {
	app := walk.App()
	// These specify the app chart sub directory for the settings file.
	app.SetOrganizationName("lab73")
	app.SetProductName("oxygen")

	// Settings file name.
	sets := &AppSetting{walk.NewIniFileSettings("settings.ini")}
	logger.Println(sets.FilePath())

	// All settings marked as expiring will expire after this duration w/o use.
	// This applies to all widgets settings.
	//settings.SetExpireDuration(time.Hour * 24 * 30 * 3)

	if err := sets.Load(); err != nil {
		logger.Fatal(err)
	}

	app.SetSettings(sets)

	if _, err := os.Stat(sets.FilePath()); os.IsNotExist(err) {
		if err := sets.Save(); err != nil {
			logger.Fatal(err)
		}
	}

	mw := new(AppMainWindow)

	config := &ModelConfig{
		appFolderPath: filepath.Dir(sets.FilePath()),
		mainWindow:    mw,
	}

	mw.model = NewModel(config)

	sets.GetJson("SERIALS", &mw.model.party.serials)

	if err := newMainwindow(mw).Create(); err != nil {
		logger.Panic(err)
	}

	mw.initialize()

	worker := NewWorker(mw.model) // поток считывания из компорта
	mw.chart = worker.chart       // запомнить ссылку на чарт
	wgWorker := sync.WaitGroup{}
	go func() {
		wgWorker.Add(1)
		worker.Run()
		wgWorker.Done()
	}() // запустить воркер
	mw.Run()            // выполнить оконную продцедуру
	worker.Stop()       // попросить воркер остановиться
	wgWorker.Wait()     // дождаться окончания воркера
	mw.model.db.Close() // закрыть базу данных

	// сохранить настройки
	sets.PutJson("SERIALS", mw.model.party.serials)
	//sets.PutTime("PARTY_ID", mw.model.party.chart.CreatedAt)

	if err := sets.Save(); err != nil {
		logger.Fatal(err)
	}
}
