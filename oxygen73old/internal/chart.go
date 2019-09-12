package internal

import (
	"bytes"
	"encoding/binary"
	"github.com/boltdb/bolt"
	"time"
)

type Chart struct {
	CreatedAt time.Time
}

func (x *Chart) rootKey() []byte {
	return itob(x.CreatedAt.UnixNano())
}

func (x *Chart) storeTimeValue(tx *bolt.Tx, timeValue TimeValue) (err error) {
	buck, err := GetBucketFromTx(tx, true, [][]byte{
		[]byte("charts"),
		x.rootKey(),
		{timeValue.N},
	})
	if err != nil {
		return err
	}
	return timeValue.store(buck)
}

func (x *Chart) getUpdatedAt(tx *bolt.Tx) (updatedAt time.Time, err error) {
	updatedAt = x.CreatedAt
	buckParties := tx.Bucket([]byte("charts"))
	if buckParties == nil {
		return
	}
	buckRoot := buckParties.Bucket(x.rootKey())
	if buckRoot == nil {
		return
	}

	for i := byte(0); i < 52; i++ {
		buck := buckRoot.Bucket([]byte{i})
		if buck == nil {
			continue
		}
		c := buck.Cursor()

		for k, v := c.First(); k != nil && v != nil; k, v = c.Next() {

			buf := bytes.NewReader(k) // ключ - наносекунды даты-времени
			var nanos int64
			if err = binary.Read(buf, binary.LittleEndian, &nanos); err != nil {
				return
			}
			if updatedAt.UnixNano() < nanos {
				updatedAt = time.Unix(0, nanos)
			}
		}
	}
	return
}

func (x *Chart) restoreSeries(tx *bolt.Tx) (r []TimeValue, err error) {
	buckParties := tx.Bucket([]byte("charts"))
	if buckParties == nil {
		return
	}
	buckRoot := buckParties.Bucket(x.rootKey())
	if buckRoot == nil {
		return
	}

	for i := byte(0); i < 52; i++ {
		buck := buckRoot.Bucket([]byte{i})
		if buck == nil {
			continue
		}
		c := buck.Cursor()
		for k, v := c.First(); k != nil && v != nil; k, v = c.Next() {
			timeValue := TimeValue{N: i}
			if err = timeValue.restore(k, v); err != nil {
				return
			}
			r = append(r, timeValue)

		}
	}
	return
}
