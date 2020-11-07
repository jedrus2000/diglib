package storage

import (
	"encoding/json"
	"fmt"
	bolt "go.etcd.io/bbolt"
)

const (
	DEFAULT_COLLECTION_NAME = "default"
)

type Storage struct {
	db *bolt.DB
}

func (storage *Storage) Open() {
	var err error
	storage.db, err = bolt.Open("diglib.db", 0600, nil)
	if err != nil {
		panic(err)
	}
	err = storage.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(DEFAULT_COLLECTION_NAME))
		if err != nil {
			return fmt.Errorf("create : %s", err)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func (storage *Storage) Close() {
	err := storage.db.Close()
	if err != nil {
		fmt.Printf("Error while closing storage: %s", err)
	}
}

func (storage *Storage) InsertItem(item *Item) {
	err := storage.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(DEFAULT_COLLECTION_NAME))
		buf, err := json.Marshal(item)
		if err != nil {
			return err
		}
		err = b.Put([]byte(item.Guid), buf)
		return err
	})
	if err != nil {
		panic(err)
	}
}

func (storage *Storage) InsertItems(items *[]Item) {
	err := storage.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(DEFAULT_COLLECTION_NAME))
		for _, item := range *items {
			buf, err := json.Marshal(item)
			if err != nil {
				return err
			}
			err = b.Put([]byte(item.Guid), buf)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func (storage *Storage) ForEach(fn func(item *Item)) {
	err := storage.db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(DEFAULT_COLLECTION_NAME))
		err := b.ForEach(func(k, v []byte) error {
			var item Item
			err := json.Unmarshal(v, &item)
			fn(&item)
			return err
		})
		return err
	})
	if err != nil {
		panic(err)
	}
}
func (storage *Storage) Find(guid string) Item {
	var item Item
	err := storage.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(DEFAULT_COLLECTION_NAME)).Cursor()
		k, v := c.Seek([]byte(guid))
		if k != nil {
			err := json.Unmarshal(v, &item)
			if err != nil {
				panic(err)
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return item
}
