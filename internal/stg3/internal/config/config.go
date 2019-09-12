package config

import (
	"github.com/fpawel/goutils/serial/modbus"
	"github.com/go-ini/ini"
	"sort"
	"strconv"
	"sync"
)

type Config struct {
	ini      *ini.File
	mu       *sync.Mutex
	fileName string
}

func New(fileName string) (x Config) {
	iniF, err := ini.LooseLoad(fileName)
	if err != nil {
		panic(err)
	}
	return Config{
		fileName: fileName,
		mu:       new(sync.Mutex),
		ini:      iniF,
	}
}

func (x Config) Close() error {
	x.mu.Lock()
	defer x.mu.Unlock()
	return x.ini.SaveToIndent(x.fileName, "    ")
}

func (x Config) PortName() string {
	x.mu.Lock()
	defer x.mu.Unlock()
	return x.ini.Section("comport").Key("name").Value()
}

func (x Config) SetPortName(v string) {
	x.mu.Lock()
	x.ini.Section("comport").Key("name").SetValue(v)
	x.mu.Unlock()
}

func (x Config) OrderOfAddr(addr modbus.Addr) int {
	x.mu.Lock()
	defer x.mu.Unlock()

	xs := x.addresses()
	for i, a := range xs {
		if a == addr {
			return i
		}
	}
	return -1
}

func (x Config) RemoveProductAddr(addr modbus.Addr) bool {
	x.mu.Lock()
	defer x.mu.Unlock()

	xs := x.addresses()
	for i, a := range xs {
		if a == addr {
			x.productsSect().DeleteKey(strconv.Itoa(i))
			return true
		}
	}
	return false
}

func (x Config) AddProduct() int {
	x.mu.Lock()
	defer x.mu.Unlock()
	n := x.productsCount()
	x.setAddrAt(n, 0)
	return n
}

func (x Config) AddProductAddr(addr modbus.Addr) (int, bool) {
	x.mu.Lock()
	defer x.mu.Unlock()

	xs := x.addresses()
	for i, a := range xs {
		if a == addr {
			return i, false
		}
	}

	xs = append(x.addresses(), addr)
	sort.Slice(xs, func(i, j int) bool {
		return xs[i] < xs[j]
	})
	x.ini.DeleteSection("Products")

	n := -1
	for i, a := range xs {
		x.setAddrAt(i, a)
		if a == addr {
			n = i
		}
	}
	if n == -1 {
		panic("unexpected: not added")
	}
	return n, true

}

func (x Config) SetAddrAt(n int, addr modbus.Addr) {
	x.mu.Lock()
	x.productAddrKeyAt(n).SetValue(strconv.Itoa(int(addr)))
	x.mu.Unlock()
}

func (x Config) AddrAt(n int) (addr modbus.Addr) {
	x.mu.Lock()
	defer x.mu.Unlock()

	v, _ := x.productAddrKeyAt(n).Int()
	return modbus.Addr(v)
}

func (x Config) Addresses() (result []modbus.Addr) {
	x.mu.Lock()
	defer x.mu.Unlock()
	return x.addresses()
}

func (x Config) ProductsCount() int {
	x.mu.Lock()
	defer x.mu.Unlock()
	return x.productsCount()
}

func (x Config) productsCount() (count int) {
	for _, k := range x.productsSect().Keys() {
		if n, _ := strconv.Atoi(k.Name()); n > count {
			count = n
		}
	}
	return
}

func (x Config) productAddrKeyAt(n int) *ini.Key {
	return x.productsSect().Key(strconv.Itoa(n))
}

func (x Config) productsSect() *ini.Section {
	return x.ini.Section("Products")
}

func (x Config) setAddrAt(n int, addr modbus.Addr) {
	x.productAddrKeyAt(n).SetValue(strconv.Itoa(int(addr)))
}

func (x Config) addresses() (result []modbus.Addr) {
	for _, k := range x.productsSect().Keys() {
		n, err := strconv.Atoi(k.Name())
		if err != nil {
			panic(err)
		}
		if len(result) < n+1 {
			xs := make([]modbus.Addr, n+1)
			copy(xs, result)
			result = xs
		}
		v, err := k.Int()
		if err != nil {
			panic(err)
		}
		result[n] = modbus.Addr(v)
	}
	return
}
