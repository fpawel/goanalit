package main

import (
	"fmt"
	coef2 "github.com/fpawel/goanalit/internal/coef"
	kgsdum2 "github.com/fpawel/goanalit/internal/kgsdum/kgsdum"
	products2 "github.com/fpawel/goanalit/internal/kgsdum/products"
	wask62 "github.com/fpawel/goanalit/internal/kgsdum/wask6"
	"github.com/lxn/win"
	"log"
	"time"
)

type workProvider struct {
	app *App
}

func (x workProvider) CheckCanceled() error {
	if x.app.ports.Canceled() {
		return fetch.ErrorCanceled
	}
	return nil
}

func (x workProvider) SetupGas(gas kgsdum2.Gas) error {
	return x.CheckCanceled()
}

func (x workProvider) SetupTemperature(temperature kgsdum2.Temperature) error {
	return x.CheckCanceled()
}

func (x workProvider) Delay(what string, duration time.Duration) error {
	x.app.Delay(what, duration, func() {
		time.Sleep(time.Millisecond * 300)
	})
	return x.CheckCanceled()

}

func (x workProvider) WriteCoefficient(coefficient coef2.Coefficient, value float64) error {
	return x.app.ProcessEachProduct(fmt.Sprintf("запись коэффициента %d: %g", coefficient, value),
		fmt.Sprintf("%d: %g", coefficient, value), func(p products2.ProductInfo) error {
			pc := coef2.AddrCoefficient{Addr: p.Addr, Coefficient: coefficient}
			return wask62.WriteCoefficient(float32(value), pc, x.app.PortProducts())
		})
}

func (x workProvider) ReadVar(v kgsdum2.Var) error {
	return x.app.ProcessEachProduct(fmt.Sprintf("чтение по адресу %d", v), fmt.Sprintf("%d", v), func(p products2.ProductInfo) error {
		_, err := wask62.ReadVar(wask62.DeviceAddr(p.Addr), wask62.ValueAddr(v), x.app.PortProducts())
		return err
	})
}

func (x workProvider) FixTestValue(test kgsdum2.Test, varCode kgsdum2.Var, n byte) error {
	gas := kgsdum2.TestGases()[n]
	return x.app.ProcessEachProduct("считать и сохранить значение",
		fmt.Sprintf("%v, %v, %v:%d", test, varCode, gas, n), func(p products2.ProductInfo) error {
			value, err := wask62.ReadVar(wask62.DeviceAddr(p.Addr), wask62.ValueAddr(varCode), x.app.PortProducts())
			if err == nil {
				x.app.db.Update(func(tx products2.Tx) {
					product, ok := tx.GetCurrentPartyProductByProductTime(p.ProductTime)
					if !ok {
						log.Fatal("product not found in current party", p)
					}
					kgsdum2.Value{
						Product: product,
						Index:   n,
						Test:    test,
						Var:     varCode,
					}.SetValue(value)
				})
			}
			return err
		})
}

func (x workProvider) Config() kgsdum2.Config {
	return x.app.config.Work
}

func (x workProvider) Print(test kgsdum2.Test, args ...interface{}) {
	x.app.WriteLogParty(products2.TestLogRecord{
		Level: win.NIIF_INFO,
		Test:  test.String(),
		Text:  fmt.Sprint(args),
	})
}

func (x workProvider) Printf(test kgsdum2.Test, format string, args ...interface{}) {
	x.app.WriteLogParty(products2.TestLogRecord{
		Level: win.NIIF_INFO,
		Test:  test.String(),
		Text:  fmt.Sprintf(format, args),
	})
}
