package products

import (
	"github.com/boltdb/bolt"
)

var keyLastLogTime = [][]byte{[]byte("last_log_time")}

func (x DB) View(f func(tx Tx)) {
	err := x.DB.View(func(tx *bolt.Tx) error {
		f(Tx{tx, x})
		return nil

	})
	if err != nil {
		panic(err)
	}
}

// Products список продуктов в последней партии
func (x DB) Products() (xs ProductInfoList) {
	x.View(func(tx Tx) {
		xs = tx.Party().ProductsInfo()
	})
	return
}

func (x DB) Update(f func(tx Tx)) {
	err := x.DB.Update(func(tx *bolt.Tx) error {
		f(Tx{tx, x})
		return nil

	})
	if err != nil {
		panic(err)
	}
}
