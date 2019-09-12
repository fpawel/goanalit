package products

import (
	"testing"

	"github.com/boltdb/bolt"
	"os"
	"time"
)

const (
	productTypeIndex = 101
	productsCount    = 20
	tempFileName     = "test_products_db.tmp"
)

var key1 []byte = []byte{1, 2, 3, 4, 5}
var value1 []byte = []byte{6, 7, 8, 9, 10}

var db DB
var dbFilePath = os.TempDir() + "//" + tempFileName
var partyTime PartyTime
var partyTime1 = PartyTime(time.Now().Add(time.Hour * 24))

// открыть базу данных
func TestOpenDB(t *testing.T) {
	var err error
	db.DB, err = bolt.Open(dbFilePath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil && os.IsNotExist(err) {
		t.Fatal(err)
	}

	// полностью очистить базу данных
	db.DB.Update(func(tx *bolt.Tx) error {
		c := tx.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if v == nil {
				if err := c.Bucket().DeleteBucket(k); err != nil {
					t.Fatal(err)
				}
			} else {
				c.Bucket().Delete(k)
			}
		}
		return nil
	})

	// убедиться что база пуста
	db.DB.View(func(tx *bolt.Tx) error {
		c := tx.Cursor()
		if k, _ := c.First(); k != nil {
			t.Fatal("not empty")
		}
		return nil
	})
}

func TestNewParty(t *testing.T) {

	db.Update(func(tx Tx) {
		party := tx.Party()
		partyTime = party.PartyTime
		party.AddProducts(productsCount)
		party.SetProductTypeIndex(productTypeIndex)
	})

	db.View(func(tx Tx) {
		party := Party{partyTime, tx}
		if len(party.Products()) != productsCount {
			t.Errorf("products count must be %d, but %d", productsCount, len(party.Products()))
		}
		if party.ProductTypeIndex() != productTypeIndex {
			t.Errorf("productTypeIndex must be %d, but %d", productTypeIndex, party.ProductTypeIndex())
		}
		if len(tx.Parties()) != 1 {
			t.Errorf("parties count must be 1, but %d", len(tx.Parties()))
		}
	})

}

func TestNewParty1(t *testing.T) {

	db.Update(func(tx Tx) {
		party := Party{partyTime1, tx}
		party.AddProducts(productsCount)
	})

	db.View(func(tx Tx) {
		party := tx.Party()
		if !party.PartyTime.Equal(partyTime1) {
			t.Errorf("party not valid: is %v, must be %v, last is %v",
				party.PartyTime.Time().UnixNano(),
				partyTime1.Time().UnixNano(),
				tx.Party().PartyTime.Time().UnixNano())
		}

		if len(party.Products()) != productsCount {
			t.Errorf("products count must be %d, but %d", productsCount, len(party.Products()))
		}
		if len(tx.Parties()) != 2 {
			t.Errorf("parties count must be 2, but %d", len(tx.Parties()))
		}
	})
}

func TestDeleteParty1(t *testing.T) {

	// удалить пратию partyTime1
	db.Update(func(tx Tx) {
		tx.DeleteParties(func(pt PartyTime) bool {
			return pt.Equal(partyTime1)
		})
	})

	// в базе должна быть только одна пратия partyTime
	db.View(func(tx Tx) {
		party := tx.Party()

		if !party.PartyTime.Equal(partyTime) {
			t.Errorf("party not valid")
		}

		if len(tx.Parties()) != 1 {
			t.Errorf("parties count must be 1, but %d", len(tx.Parties()))
		}
	})
}

func TestUpdateProduct(t *testing.T) {

	// удалить пратию partyTime1
	db.Update(func(tx Tx) {
		p := tx.Party().Products()[0]
		p.SetSerial(100)
		p.SetAddr(10)
		tx.SetValue(p.Path(), key1, value1)

	})

	// в базе должна быть только одна пратия partyTime
	db.View(func(tx Tx) {
		p := tx.Party().Products()[0]

		if p.Addr() != 10 {
			t.Errorf("addr must be 10")
		}

		if p.Serial() != 100 {
			t.Errorf("serial must be 10")
		}

		value := tx.Value(p.Path(), key1)
		if len(value) != len(value1) {
			t.Errorf("value %v is not %v", value1, value)
		}

		for i := range value {
			if value[i] != value1[i] {
				t.Errorf("value %v is not %v", value1, value)
			}
		}

	})
}
