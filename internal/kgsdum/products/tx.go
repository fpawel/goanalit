package products

import (
	"github.com/boltdb/bolt"
	"github.com/fpawel/gutils/boltDBHelp"
	"github.com/fpawel/gutils/utils"
	"time"
)

func (x Tx) NewParty(productsCount int) Party {
	newParty := Party{PartyTime: PartyTime(time.Now()), Tx: x}
	newParty.AddProducts(productsCount)
	return newParty
}

func (x Tx) BucketRead(path [][]byte) *bolt.Bucket {
	sr := boltDBHelp.NewBucketSearcher(x.tx, false)
	sr.Find(path)
	if sr.Error() == boltDBHelp.ErrorBucketNotExist {
		return nil
	}
	if sr.Error() != nil {
		panic(sr.Error())
	}
	return sr.Bucket()
}
func (x Tx) BucketWrite(path [][]byte) *bolt.Bucket {
	sr := boltDBHelp.NewBucketSearcher(x.tx, true)
	sr.Find(path)
	if sr.Error() != nil {
		panic(sr.Error())
	}
	return sr.Bucket()
}

func (x Tx) Value(path [][]byte, key []byte) []byte {
	buck := x.BucketRead(path)
	if buck == nil {
		return nil
	}
	return buck.Get(key)
}

func (x Tx) SetValue(path [][]byte, key []byte, value []byte) {

	if value == nil {
		if err := x.BucketWrite(path).Delete(key); err != nil {
			panic(err)
		}
	} else {
		if err := x.BucketWrite(path).Put(key, value); err != nil {
			panic(err)
		}
	}
	return
}

func (x Tx) Parties() (r []Party) {
	buck := x.tx.Bucket(keyParties)
	if buck != nil {
		c := buck.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if v == nil {
				r = append(r, Party{PartyTime(time.Unix(0, utils.BytesToInt64(k))), x})
			}
		}
	}
	return
}

func (x Tx) Party() Party {
	xs := x.Parties()
	n := len(xs)
	if len(xs) == 0 {
		return Party{PartyTime(time.Now()), x}
	}
	return xs[n-1]
}

func (x Tx) GetProductByProductTime(partyTime PartyTime, productTime ProductTime) (Product, bool) {
	for _, party := range x.Parties() {
		if party.PartyTime == partyTime {
			for _, p := range party.Products() {
				if p.ProductTime == productTime {
					return p, true
				}
			}
		}
	}
	var tmp Product
	return tmp, false
}

func (x Tx) GetCurrentPartyProductByProductTime(productTime ProductTime) (Product, bool) {
	return x.GetProductByProductTime(x.Party().PartyTime, productTime)
}

func (x Tx) GetCurrentPartyProductByAddr(addr byte) (Product, bool) {
	for _, party := range x.Parties() {
		if party.PartyTime == x.Party().PartyTime {
			for _, p := range party.Products() {
				if p.Addr() == addr {
					return p, true
				}
			}
		}
	}
	var tmp Product
	return tmp, false
}

func (x Tx) DeleteParties(f func(PartyTime) bool) {

	buckParties := x.tx.Bucket(keyParties)
	if buckParties != nil {
		c := buckParties.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if v == nil {
				t := PartyTime(time.Unix(0, utils.BytesToInt64(k)))
				if f(t) {
					if err := buckParties.DeleteBucket(k); err != nil {
						panic(err)
					}
				}
			}
		}
	}
	return
}
