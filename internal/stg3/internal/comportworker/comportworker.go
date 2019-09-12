package comportworker

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fpawel/goanalit/stg3alit/stg3/internal/config"
	"github.com/fpawel/goanalit/stg3alit/stg3/internal/stg"
	"github.com/fpawel/goanalit/stg3alit/stg3/internal/tasks"
	"github.com/fpawel/goutils/procmq"
	"github.com/fpawel/goutils/serial/comport"
	"github.com/fpawel/goutils/serial/fetch"
	"github.com/fpawel/goutils/serial/modbus"
	"github.com/tarm/serial"
	"sync"
	"sync/atomic"
	"time"
)

type comportWorker struct {
	config   config.Config
	peer     *procmq.ProcessMQ
	port     *comport.Port
	scanAddr uint32
	close    uint32
	tasks    *tasks.Tasks
	wgClose  *sync.WaitGroup
}

func Run() error {

	x := &comportWorker{
		peer:    procmq.MustOpen("STG3"),
		config:  config.New("stg3.ini"),
		tasks:   new(tasks.Tasks),
		wgClose: new(sync.WaitGroup),
	}

	x.port = comport.NewPort(comport.Config{
		Serial: serial.Config{
			Baud:        9600,
			ReadTimeout: time.Millisecond,
		},
		Fetch: fetch.Config{
			ReadTimeout:     500 * time.Millisecond,
			ReadByteTimeout: 50 * time.Millisecond,
		},
	})

	x.registerRoutes()

	fmt.Println("stg3: connecting to peer...")
	if err := x.peer.Connect(); err != nil {
		panic(err)
	}
	fmt.Println("stg3: peer connected")

	x.wgClose.Add(2)

	go x.runPeer()
	go x.runComport()

	x.wgClose.Wait()

	var errs []error
	if err := x.port.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := x.config.Close(); err != nil {
		errs = append(errs, err)
	}
	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[1]
	default:
		return errors.New(errs[0].Error() + ", " + errs[1].Error())
	}
}

// registerRoutes создаётроуты, через которые приходят команды от клиентского приложения
func (x *comportWorker) registerRoutes() {
	for msg, fun := range map[string]func([]byte) interface{}{

		"ADD_PLACE": func([]byte) interface{} {
			x.config.AddProduct()
			return nil
		},

		"ADDR": func(b []byte) interface{} {
			var a struct {
				Place int
				Addr  modbus.Addr
			}
			mustUnmarshalJson(b, &a)
			x.config.SetAddrAt(a.Place, a.Addr)
			for i := range stg.Vars {
				x.tasks.Put(a.Place, i, tasks.Read, 0)
			}
			return nil
		},
	} {
		x.peer.Handle(msg, fun)
	}
}

func (x *comportWorker) runPeer() {
	defer x.wgClose.Done()

	if err := x.peer.Run(); err != nil {
		panic(err)
	}
	atomic.AddUint32(&x.close, 1)

}

func (x *comportWorker) runComport() {
	defer x.wgClose.Done()

	portName := x.config.PortName()
	if portName == "" {
		ports, _ := comport.GetAvailablePorts()
		if len(ports) > 0 {
			portName = ports[0]
		}
	}
	x.config.SetPortName(portName)
	x.peer.Send("PORT_NAME", portName)

	x.notifyProducts()
	place := -1
	for atomic.LoadUint32(&x.close) == 0 {

		if x.openPort(); !x.port.Opened() {
			continue
		}

		x.doScan()
		x.doVars()
		if place++; place > x.config.ProductsCount() {
			place = 0
		}
		x.readConcentration(place)
	}
}

func (x *comportWorker) doVars() {
	for _, v := range x.tasks.PopList() {
		addr := x.config.AddrAt(v.Place)
		if addr == 0 {
			continue
		}
		stgVar := stg.Vars[v.VarNumber]
		var err error
		if v.Action == tasks.Write {
			err = modbus.Write32Float(x.port, addr, stgVar.Cmd, v.Value)
		} else {
			v.Value, err = modbus.Read3BCD(x.port, addr, stgVar.Var)
		}

		x.notifyVar(v.Place, v.VarNumber, v.Value, err)
		if err != nil {
			x.tasks.Put(v.Place, v.VarNumber, v.Action, v.Value)
		}
	}
}

func (x *comportWorker) doScan() {
	if x.config.ProductsCount() == 0 {
		atomic.StoreUint32(&x.scanAddr, 1)
	}
	for addr := atomic.LoadUint32(&x.scanAddr); addr > 0 && addr < 128; atomic.AddUint32(&x.scanAddr, 1) {

		addr := modbus.Addr(addr)

		value, err := modbus.Read3BCD(x.port, addr, 0)
		if err == nil {
			if _, addedNew := x.config.AddProductAddr(addr); addedNew {
				x.notifyProducts()
			}
		}
		if n := x.config.OrderOfAddr(addr); n > -1 {
			x.notifyConcentration(n, value, err)
		}
		x.peer.Send("SCAN", struct{ Addr modbus.Addr }{addr})
	}
}

func (x comportWorker) openPort() {
	if x.port.Name() != x.config.PortName() {
		x.port.Close()
		x.port.SetName(x.config.PortName())
	}
	if x.port.Opened() {
		return
	}
	err := x.port.Open()
	if err != nil {
		x.peer.Send("ERROR", err.Error())
	}
}

func (x comportWorker) notifyProducts() {
	x.peer.Send("PRODUCTS", struct {
		Products []modbus.Addr
	}{x.config.Addresses()})
}

func (x *comportWorker) readConcentration(place int) {
	addr := x.config.AddrAt(place)
	if addr > 0 {
		value, err := modbus.Read3BCD(x.port, addr, 0)
		x.notifyConcentration(place, value, err)
	}
}

func (x *comportWorker) notifyConcentration(place int, value float64, err error) {
	x.peer.Send("CONCENTRATION", struct {
		Place int
		Value float64
		Error string
	}{place, value, fmtErr(err)})
}

func (x *comportWorker) notifyVar(place, varNumber int, value float64, err error) {
	x.peer.Send("VAR", struct {
		Place, Var int
		Value      float64
		Error      string
	}{place, varNumber, value, fmtErr(err)})
}

func fmtErr(e error) string {
	if e != nil {
		return e.Error()
	}
	return ""
}

func mustUnmarshalJson(b []byte, v interface{}) {
	if err := json.Unmarshal(b, v); err != nil {
		panic(err.Error() + ": " + string(b))
	}
}
