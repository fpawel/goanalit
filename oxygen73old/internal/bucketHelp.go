package internal

import (
	"fmt"
	"github.com/boltdb/bolt"
)

func GetBucketFromTx(tx *bolt.Tx, canCreate bool, path [][]byte) (*bolt.Bucket, error) {
	var buck *bolt.Bucket
	var err error
	if canCreate {
		buck, err = tx.CreateBucketIfNotExists(path[0])
		if err != nil {
			return nil, err
		}
	} else {
		buck = tx.Bucket(path[0])
		if buck == nil {
			err = fmt.Errorf("bucket not exists %s, %v", string(path[0]), path[0])
			return nil, err
		}
	}
	return GetBucketFromBucket(buck, canCreate, path[1:])
}

func GetBucketFromBucket(buck *bolt.Bucket, canCreate bool, path [][]byte) (*bolt.Bucket, error) {

	var err error
	for i, k := range path {

		if canCreate {
			buck, err = buck.CreateBucketIfNotExists(k)
			if err != nil {
				return nil, err
			}
		} else {
			buck = buck.Bucket(k)
			if buck == nil {
				s := ""
				for _, a := range path {
					s += fmt.Sprintf("%s %v, ", string(a), a)
				}
				return nil, fmt.Errorf("bucket not exists %d, %s", i, s)

			}
		}
	}
	return buck, nil
}
