package products

import (
	"fmt"
	b "github.com/fpawel/gutils/boltDBHelp"
	u "github.com/fpawel/gutils/utils"
	"time"
)

func (x PartyTime) Path() [][]byte {
	return [][]byte{keyParties, TimeToKey(x.Time())}
}

func (x PartyTime) Time() time.Time {
	return time.Time(x)
}

func (x Party) Path() [][]byte {
	return x.PartyTime.Path()

}

func (x Party) String() string {
	t := time.Time(x.PartyTime)
	return fmt.Sprintf("%s, %d приборов", t.Format("15:04"), len(x.Products()))
}

func (x Party) ForEachProduct(f func(Product)) {
	sr := b.NewBucketSearcher(x.Tx.tx, false)
	sr.Find(append(x.Path(), keyProducts))
	if sr.Error() == b.ErrorBucketNotExist {
		return
	}
	if sr.Error() != nil {
		panic(sr.Error())
	}
	c := sr.Bucket().Cursor()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		if v == nil {
			f(Product{ProductTime(time.Unix(0, u.BytesToInt64(k))), x})
		}
	}
	return
}

func (x Party) Products() (r []Product) {
	x.ForEachProduct(func(product Product) {
		r = append(r, product)
	})
	return
}

func (x Party) Info() PartyInfo {
	return PartyInfo{
		PartyTime:        x.PartyTime,
		ProductTypeIndex: x.ProductTypeIndex(),
	}
}

func (x Party) ProductsInfo() (r ProductInfoList) {
	for row, p := range x.Products() {
		i := p.Info()
		i.Row = row
		i.Party = x.Info()
		r = append(r, i)
	}
	return
}

func (x Party) Test(s string) DBPath {
	return pathInfo{append(x.Path(), []byte("tests"), []byte(s))}
}

func (x Party) AddProducts(count int) {
	addrs := make(map[byte]struct{})
	tt := time.Now()
	x.ForEachProduct(func(product Product) {
		addrs[product.Addr()] = struct{}{}
	})
	for np := 0; np < count; np++ {
		var addr byte
		for i := byte(1); i <= 255; i++ {
			if _, ok := addrs[i]; !ok {
				addr = i
				break
			}
		}
		addrs[addr] = struct{}{}

		Product{
			ProductTime: ProductTime(tt.Add(time.Millisecond * time.Duration(np))),
			Party:       x,
		}.SetAddr(addr)
	}
}

func (x Party) AddProduct(addr byte, serial uint64) {
	partyTime := x.PartyTime.Time()
	productTime := partyTime.Add(time.Millisecond * time.Duration(len(x.Products())+1))
	p := Product{
		ProductTime: ProductTime(productTime),
		Party:       x,
	}
	p.SetAddr(addr)
	p.SetSerial(serial)
}

func (x Party) DeleteProducts(ns []int) {
	xs := make(map[int]struct{})
	for _, n := range ns {
		xs[n] = struct{}{}
	}

	sr := b.NewBucketSearcher(x.Tx.tx, false)
	sr.Find(append(x.Path(), keyProducts))
	if sr.Error() == b.ErrorBucketNotExist {
		return
	}
	if sr.Error() != nil {
		panic(sr.Error())
	}
	c := sr.Bucket().Cursor()
	n := 0
	for k, v := c.First(); k != nil; k, v = c.Next() {
		if v == nil {
			if _, f := xs[n]; f {
				if err := sr.Bucket().DeleteBucket(k); err != nil {
					panic(err)
				}
			}
			n++
		}
	}
}

func (x Party) ProductTypeIndex() (r int) {
	b := x.Tx.Value(x.Path(), []byte("PRODUCTS_TYPE"))
	if len(b) > 0 {
		r = int(b[0])
	}
	return
}

func (x Party) SetProductTypeIndex(r int) {
	x.Tx.SetValue(x.Path(), []byte("PRODUCTS_TYPE"), []byte{byte(r)})
}

func (x Party) WriteLog(r TestLogRecord) []byte {
	return x.Tx.WriteTestLog(x, r)
}
