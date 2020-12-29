package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	badger "github.com/dgraph-io/badger/v2"
)

const (
	DEFAULT_COLLECTION_NAME = "default"
)

type Storage struct {
	db *badger.DB
}

func (storage *Storage) Open() {
	var err error
	dstPath := filepath.Join("database")
	options := badger.DefaultOptions(dstPath)
	options.Truncate = true
	storage.db, err = badger.Open(options)
	if err != nil {
		panic(err)
	}
}

func (storage *Storage) Close() {
	println("Closing DB...")
	err := storage.db.Close()
	if err != nil {
		fmt.Printf("Error while closing storage: %s", err)
	}
}

func (storage *Storage) SaveItem(item *Item, overwrite bool) {
	err := storage.db.Update(func(txn *badger.Txn) error {
		if i, _ := txn.Get([]byte(item.Guid)); (i == nil) || (overwrite == true) {
			buf, err := json.Marshal(item)
			if err != nil {
				return err
			}
			err = txn.Set([]byte(item.Guid), buf)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func (storage *Storage) SaveItems(items *[]Item, overwrite bool) {
	err := storage.db.Update(func(txn *badger.Txn) error {
		for _, item := range *items {
			if i, _ := txn.Get([]byte(item.Guid)); (i == nil) || (overwrite == true) {
				buf, err := json.Marshal(item)
				if err != nil {
					return err
				}
				err = txn.Set([]byte(item.Guid), buf)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func (storage *Storage) ForEach(fn func(item *Item)) {
	err := storage.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			v := it.Item()
			// k := v.Key()
			err := v.Value(func(data []byte) error {
				var item Item
				err := json.Unmarshal(data, &item)
				fn(&item)
				return err
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func (storage *Storage) Read(guid string) (Item, error) {
	var item Item
	var err error = nil

	err = storage.db.View(func(txn *badger.Txn) error {
		i, err := txn.Get([]byte(guid))
		if err == nil {
			err := i.Value(func(data []byte) error {
				return json.Unmarshal(data, &item)
			})
			if err != nil {
				panic(err)
			}
		} else {
			err = errors.New("item not found")
		}
		return err
	})
	return item, err
}

/*
	storage.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(guid)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()
			err := item.Value(func(v []byte) error {
				fmt.Printf("key=%s, value=%s\n", k, v)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})


	err = storage.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(DEFAULT_COLLECTION_NAME)).Cursor()
		k, v := c.Seek([]byte(guid))
		if (k != nil) && (string(k) == guid) {
			err := json.Unmarshal(v, &item)
			if err != nil {
				return err
			}
		} else {
			return errors.New("item not found")
		}
		return nil
	})
	return item, err
}
*/
