package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/daf.v0/internal/data"
	"github.com/fpawel/daf.v0/internal/viewmodel"
	"github.com/fpawel/gohelp"
	"github.com/hako/durafmt"
	"github.com/lxn/walk"
	"github.com/powerman/structlog"
	"time"
)

func isFailWork(err error) bool {
	return err != nil && !isDeviceError(err)
}

func isDeviceError(err error) bool {
	return merry.Is(err, comm.Err) || merry.Is(err, context.DeadlineExceeded)
}

type Product = viewmodel.DafProductViewModel

func doForEachOkProduct(work func(p *Product) error) error {
	if len(prodsMdl.OkProducts()) == 0 {
		return ErrNoOkProducts
	}
	for _, p := range prodsMdl.OkProducts() {
		if err := work(p); isFailWork(err) {
			return err
		}
	}
	return nil
}

func dafSendCmdToEachOkProduct(cmd modbus.DevCmd, arg float64) error {

	dafMainWindow.SetWorkStatus(walk.RGB(0, 0, 128),
		fmt.Sprintf("отправка команды %X, %v", cmd, arg))

	if cmd == 5 {
		_, err := portDaf.Write(log, ctxApp, modbus.NewWrite32BCDRequest(0, 0x10, cmd, arg).Bytes())
		return err
	}

	return doForEachOkProduct(func(p *Product) error {
		return dafSendCmdToPlace(p.Place, cmd, arg)
	})
}

func blowGas(gas data.Gas) error {
	if err := switchGas(gas); err != nil {
		if !dafMainWindow.IgnoreErrorPrompt(fmt.Sprintf("газовый блок: %d", gas), err) {
			return err
		}
	}
	t := 5 * time.Minute
	if gas == 1 {
		t = 10 * time.Minute
	}
	//
	t = time.Second * 20
	//
	return delay(fmt.Sprintf("продувка ПГС%d", gas), t)
}

func dafSetupCurrent() error {
	setCurrentWorkName("настройка токового выхода")

	if err := dafSendCmdToEachOkProduct(0xB, 1); err != nil {
		return err
	}

	sleep(5 * time.Second)

	dafMainWindow.SetWorkStatus(ColorNavy, "корректировка тока 4 мА")

	if err := doForEachOkProduct(func(p *Product) error {
		v, err := EN6408Read(p.Place) // считать тока со стенда
		if err != nil {
			return err
		}
		return dafSendCmdToPlace(p.Place, 9, v.Current)
	}); err != nil {
		return err
	}

	if err := dafSendCmdToEachOkProduct(0xC, 1); err != nil {
		return err
	}

	sleep(5 * time.Second)

	dafMainWindow.SetWorkStatus(ColorNavy, "корректировка тока 20 мА")
	if err := doForEachOkProduct(func(p *Product) error {
		v, err := EN6408Read(p.Place)
		if err != nil {
			return err
		}
		return dafSendCmdToPlace(p.Place, 0xA, v.Current)
	}); err != nil {
		return err
	}
	return nil
}

func dafSetupThresholdTest() error {

	setCurrentWorkName("установка порогов для настройки")

	party := data.GetLastParty()

	if err := dafSendCmdToEachOkProduct(0x30, party.Threshold1Test); err != nil {
		return err
	}
	if err := dafSendCmdToEachOkProduct(0x31, party.Threshold2Test); err != nil {
		return err
	}

	return nil
}

func dafAdjust() error {

	party := data.GetLastParty()

	setCurrentWorkName("корректировка нулевых показаний")

	defer func() {
		_ = switchGas(0)
	}()

	if err := blowGas(1); err != nil {
		return err
	}
	if err := dafSendCmdToEachOkProduct(0x32, party.Pgs1); err != nil {
		return err
	}

	setCurrentWorkName("корректировка чувствительности")

	if err := blowGas(4); err != nil {
		return err
	}
	if err := dafSendCmdToEachOkProduct(0x33, party.Pgs4); err != nil {
		return err
	}
	return nil
}

func dafTestMeasureRange() error {

	defer func() {
		_ = switchGas(0)
	}()

	setCurrentWorkName("проверка диапазона измерений")

	for n, gas := range []data.Gas{1, 2, 3, 4, 3, 1} {

		what := fmt.Sprintf("проверка диапазона измерений: ПГС%d, точка %d", gas, n+1)

		dafMainWindow.SetWorkStatus(ColorNavy, what+": продувка газа")

		if err := blowGas(gas); err != nil {
			return err
		}

		dafMainWindow.SetWorkStatus(ColorNavy, what+": опрос и сохранение данных")

		if err := doForEachOkProduct(func(p *Product) error {

			dv, err := dafReadAtPlace(p.Place)
			if isFailWork(err) {
				return nil
			}
			v, err := EN6408Read(p.Place)
			if err != nil {
				return nil
			}

			data.DBxProducts.MustExec(
				`DELETE FROM product_value WHERE product_id = ? AND work_index = ?`,
				p.ProductID, n)

			value := data.ProductValue{
				ProductID:     p.ProductID,
				Gas:           gas,
				CreatedAt:     time.Now(),
				WorkIndex:     n,
				Concentration: dv.Concentration,
				Current:       v.Current,
				Threshold1:    v.Threshold1,
				Threshold2:    v.Threshold2,
				Mode:          dv.Mode,
				FailureCode:   dv.Failure,
			}

			structlog.New().Info("сохранение для паспорта",
				"место", p.Place,
				"адрес", p.Addr,
				"заводской_номер", p.Serial,
				"значение", fmt.Sprintf("%+v", value),
			)

			if err := data.DBProducts.Save(&value); err != nil {
				panic(err)
			}

			return nil

		}); err != nil {
			return err
		}
	}

	if err := blowGas(1); err != nil {
		return err
	}

	return nil
}

func dafTestStability() error {
	if err := blowGas(3); err != nil {
		return err
	}
	return nil
}

func dafSetupMain() error {
	for _, f := range []func() error{dafSetupCurrent, dafSetupThresholdTest, dafAdjust, dafTestMeasureRange} {
		if err := f(); err != nil {
			return err
		}
	}
	if data.GetLastParty().Type == 0 {
		return nil
	}

	setCurrentWorkName("проверка HART протокола")
	return doForEachOkProduct(testHart)
}

func interrogateProducts() error {
	currentWorkName = ""
	for {
		if len(prodsMdl.OkProducts()) == 0 {
			return errors.New("не выбрано ни одной строки в таблице приборов текущей партии")
		}
		for _, p := range prodsMdl.OkProducts() {
			if _, err := EN6408Read(p.Place); err != nil {
				return err
			}
			if _, err := dafReadAtPlace(p.Place); isFailWork(err) {
				return err
			}
		}
	}
}

func sleep(t time.Duration) {
	log := gohelp.LogPrependSuffixKeys(log, "duration", durafmt.Parse(t))

	log.Info("начало паузы", structlog.KeyTime, now())
	defer func() {
		log.Info("окончание паузы", structlog.KeyTime, now())
	}()
	timer := time.NewTimer(t)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			return
		case <-ctxApp.Done():
			return
		}
	}
}

func delay(what string, total time.Duration) error {

	log := gohelp.LogPrependSuffixKeys(log, "duration", durafmt.Parse(total), "what", what)

	originalComportContext := ctxApp
	ctxDelay, doSkipDelay := context.WithTimeout(ctxApp, total)
	ctxApp = ctxDelay
	defer func() {
		ctxApp = originalComportContext
		dafMainWindow.DelayHelp.Hide()
	}()

	startMoment := time.Now()

	skipDelay = func() {
		doSkipDelay()
		log.Warn("задержка прервана", structlog.KeyTime, now())
	}
	dafMainWindow.DelayHelp.Show(what, total)

	log.Info("начало задержки", structlog.KeyTime, now())
	defer func() {
		log.Info("окончание задержки",
			structlog.KeyTime, now(),
			"elapsed", durafmt.Parse(time.Since(startMoment)),
		)
	}()

	for {
		for _, p := range prodsMdl.OkProducts() {
			_, err := dafReadAtPlace(p.Place)
			if ctxDelay.Err() != nil {
				return nil
			}
			if isFailWork(err) {
				return err
			}
		}

		timer := time.NewTimer(5 * time.Second)
		for {
			select {
			case <-timer.C:
				break
			case <-ctxDelay.Done():
				timer.Stop()
				return nil
			}
		}

	}
}

func onPlaceConnectionError(place int, err error) {
	p := prodsMdl.ProductAt(place)
	prodsMdl.SetConnectionErrorAt(place, err)
	if currentWorkName != "" {
		data.WriteProductError(p.ProductID, currentWorkName, err)
	}
}

func setCurrentWorkName(workName string) {
	data.DBxProducts.MustExec(
		`DELETE FROM product_entry 
WHERE work_name = ? 
  AND product_id IN (SELECT product_id FROM last_party_products)`, workName)
	currentWorkName = workName
}
