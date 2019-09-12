package main

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm/modbus"
	"github.com/fpawel/daf.v0/internal/data"
	"github.com/fpawel/daf.v0/internal/viewmodel"
	"github.com/fpawel/gohelp"
	"github.com/powerman/structlog"
	"github.com/sirupsen/logrus"
)

type EN6408ConnectionLine int

const (
	EN6408Disconnect EN6408ConnectionLine = iota
	EN6408ConnectRS485
	EN6408ConnectHart
)

var (
	ErrEN6408   = merry.New("стенд 6408")
	ErrGasBlock = merry.New("газовый блок")
)

func EN6408SetConnectionLine(place int, connLine EN6408ConnectionLine) error {
	req := modbus.Request{
		Addr:     0x20,
		ProtoCmd: 0x10,
		Data:     []byte{0, byte(place), 0, 1, 2, 0, byte(connLine)},
	}

	_, err := req.GetResponse(
		gohelp.LogPrependSuffixKeys(log, "ЭН6408", "установка линии связи", "линия", connLine, "место", place+1),
		ctxApp,
		portDaf, nil)
	if err != nil {
		err = ErrEN6408.Appendf("место %d: установка линии связи %d: %+v", place+1, connLine, err)
	}
	return err
}

func EN6408Read(place int) (*viewmodel.DafValue6408, error) {

	prodsMdl.SetInterrogatePlace(place)
	defer func() {
		prodsMdl.SetInterrogatePlace(-1)
	}()

	v := new(viewmodel.DafValue6408)
	addr := prodsMdl.ProductAt(place).Addr
	_, err := modbus.Read3(gohelp.LogPrependSuffixKeys(log, "ЭН6408", "опрос", "место", place+1), ctxApp,
		portDaf, 32, modbus.Var(addr-1)*2, 2, func(_, response []byte) (string, error) {
			b := response[3:]
			v.Current = (float64(b[0])*256 + float64(b[1])) / 100
			v.Threshold1 = b[3]&1 == 0
			v.Threshold2 = b[3]&2 == 0
			return fmt.Sprintf("%+v", *v), nil
		})
	if err != nil {
		return nil, ErrEN6408.Appendf("опрос места %d: %+v", place+1, err)
	}
	prodsMdl.Set6408Value(place, *v)
	return v, nil
}

func switchGas(gas data.Gas) error {

	req := modbus.Request{
		Addr:     33,
		ProtoCmd: 0x10,
		Data:     []byte{0, 32, 0, 1, 2, 0, byte(gas)},
	}
	if _, err := req.GetResponse(
		gohelp.LogPrependSuffixKeys(log, "газовый_блок", gas), ctxApp,
		portDaf, nil); err != nil {
		return ErrGasBlock.Appendf("газовый блок: клапан %d: %v", gas, err)
	}
	return nil
}

func dafReadAtPlace(place int) (v viewmodel.DafValue, err error) {

	product := prodsMdl.ProductAt(place)
	addr := product.Addr
	log := withProductAtPlace(structlog.New(), place)

	prodsMdl.SetInterrogatePlace(place)
	defer func() {
		prodsMdl.SetInterrogatePlace(-1)
	}()

	for _, x := range []struct {
		var3 modbus.Var
		p    *float64
	}{
		{0x00, &v.Concentration},
		{0x1C, &v.Threshold1},
		{0x1E, &v.Threshold2},
		{0x20, &v.Failure},
		{0x36, &v.Version},
		{0x3A, &v.VersionID},
		{0x32, &v.Gas},
	} {
		if *x.p, err = modbus.Read3BCD(log, ctxApp, portDaf, addr, x.var3); err != nil {
			break
		}
	}
	if err == nil {
		v.Mode, err = modbus.Read3UInt16(log, ctxApp, portDaf, addr, 0x23)
	}

	if err == nil {
		prodsMdl.SetDafValue(place, v)
	}
	if isDeviceError(err) {
		onPlaceConnectionError(place, err)
	}
	return
}

func dafSendCmdToPlace(place int, cmd modbus.DevCmd, arg float64) error {
	prodsMdl.SetInterrogatePlace(place)
	defer func() {
		prodsMdl.SetInterrogatePlace(-1)
	}()

	addr := prodsMdl.ProductAt(place).Addr

	log := withProductAtPlace(structlog.New(), place)

	err := modbus.Write32(log, ctxApp, portDaf, addr, 0x10, cmd, arg)
	if err == nil {
		logrus.Infof("ДАФ №%d, адрес %d: запись в 32-ой регистр %X, %v", place+1, addr, cmd, arg)
		prodsMdl.SetConnectionOkAt(place)
		return nil
	}
	if isDeviceError(err) {
		onPlaceConnectionError(place, err)
	}
	return err
}
