package products

import (
	"github.com/boltdb/bolt"
	"github.com/fpawel/gutils/utils"
	"time"
)

type DB struct {
	DB *bolt.DB
}

type Tx struct {
	tx *bolt.Tx
	db DB
}

type PartyTime time.Time
type ProductTime time.Time

type Party struct {
	PartyTime PartyTime
	Tx        Tx
}

type Product struct {
	ProductTime ProductTime
	Party       Party
}

type PartyInfo struct {
	PartyTime        PartyTime
	ProductTypeIndex int
}

type ProductInfo struct {
	Party       PartyInfo
	ProductTime ProductTime
	Row         int
	Addr        byte
	Serial      uint64
}

type ProductInfoList []ProductInfo

type DBPath interface {
	Path() [][]byte
}

type LogRecord struct {
	Path  [][]byte
	Level int
	Text  string
}

type TestLogRecord struct {
	Test    string
	TimeKey []byte
	Level   int
	Text    string
}

type pathInfo struct {
	path [][]byte
}

type Logs map[time.Time]*LogRecord

var keyParties = []byte("parties")
var keyProducts = []byte("products")
var keyProductAddr = []byte("addr")
var keyProductSerial = []byte("serial")

func TestsPath(p DBPath, what string) DBPath {
	return pathInfo{append(p.Path(), []byte("tests"), []byte(what))}
}

func (x ProductInfoList) ProductRowByAddress(addr byte) int {
	for row, p := range x {
		if addr == p.Addr {
			return row
		}
	}
	return -1
}

func (x ProductInfoList) ProductByAddress(addr byte) (ProductInfo, bool) {
	for _, p := range x {
		if addr == p.Addr {
			return p, true
		}
	}
	return ProductInfo{}, false
}

func TimeToKey(t time.Time) []byte {
	return utils.Int64ToBytes(t.UnixNano())
}

func KeyToTime(k []byte) time.Time {
	return time.Unix(0, utils.BytesToInt64(k))
}

func (x pathInfo) Path() [][]byte {
	return x.path
}

func Test(s string) DBPath {
	return pathInfo{[][]byte{[]byte("tests"), []byte(s)}}
}

func (x PartyTime) Equal(y PartyTime) bool {
	return x.Time().UnixNano() == y.Time().UnixNano()
}
