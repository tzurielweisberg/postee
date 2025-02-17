package dbservice

import (
	"os"
	"strconv"

	hookDbService "github.com/tzurielweisberg/postee/v2/dbservice"
	bolt "go.etcd.io/bbolt"
)

func GetPlgnStats() (r map[string]int, err error) {
	r = make(map[string]int)

	var DbPath string
	if len(os.Getenv("PATH_TO_DB")) > 0 {
		DbPath = os.Getenv("PATH_TO_DB")
	} else {
		DbPath = hookDbService.DbPath
	}

	db, err := bolt.Open(DbPath, 0444, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(hookDbService.DbBucketActionStats))
		if bucket == nil {
			return nil //no bucket - empty stats will be returned
		}

		c := bucket.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			cnt, err := strconv.Atoi(string(v[:]))
			if err != nil {
				return err
			}

			r[string(k[:])] = cnt
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return r, nil
}
