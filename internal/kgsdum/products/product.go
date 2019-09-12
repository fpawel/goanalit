package products

import (
	"encoding/binary"
	"time"
)

func (x ProductTime) Time() time.Time {
	return time.Time(x)
}

func (x ProductTime) Path(p PartyTime) [][]byte {
	return append(p.Path(), keyProducts, TimeToKey(x.Time()))
}

func (x Product) Path() [][]byte {
	return x.ProductTime.Path(x.Party.PartyTime)
}

func (x Product) info() ProductInfo {
	return ProductInfo{
		ProductTime: x.ProductTime,
		Addr:        x.Addr(),
		Serial:      x.Serial(),
	}
}

func (x Product) Info() (result ProductInfo) {
	result = x.info()
	for i, p := range x.Party.Products() {
		if p.ProductTime == x.ProductTime {
			result.Row = i
			break
		}
	}
	result.Party = x.Party.Info()
	return
}

func (x Product) Addr() byte {
	b := x.Party.Tx.Value(x.Path(), keyProductAddr)
	if len(b) == 0 {
		return 0
	}
	return b[0]
}

func (x Product) Serial() uint64 {
	b := x.Party.Tx.Value(x.Path(), keyProductSerial)
	if len(b) < 8 {
		return 0
	}
	return binary.BigEndian.Uint64(b)
}

func (x Product) SetAddr(b byte) {
	x.Party.Tx.SetValue(x.Path(), keyProductAddr, []byte{b})
}

func (x Product) SetSerial(v uint64) {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	x.Party.Tx.SetValue(x.Path(), keyProductSerial, b)
}

func (x Product) Test(s string) DBPath {
	return TestsPath(x, s)
}

func (x Product) WriteLog(r TestLogRecord) []byte {
	return x.Party.Tx.WriteTestLog(x, r)
}
