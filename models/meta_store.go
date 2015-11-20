package models

import (
	"encoding/json"

	"errors"
	"github.com/boltdb/bolt"
	"github.com/spf13/viper"
)

var MStore MetaStore

type MetaStore interface {
	Close() error
	FindChannelByName(string) (*Channel, bool)
	FindChannels() ([]*Channel, error)
	UpdateChannel(*Channel) error
	InsertChannel(*Channel) error
	DeleteChannel(*Channel) error
}

type boltStore struct {
	db *bolt.DB
}

func (b *boltStore) Close() error {
	return b.db.Close()
}

func (b *boltStore) FindChannels() ([]*Channel, error) {
	chs := make([]*Channel, 0)

	err := b.db.View(func(tx *bolt.Tx) error {
		buc := tx.Bucket([]byte("channels"))
		if buc == nil {
			return errors.New("bucket does not exist")
		}

		cur := buc.Cursor()

		for k, v := cur.First(); k != nil; k, v = cur.Next() {
			ch := &Channel{}
			if err := json.Unmarshal(v, ch); err != nil {
				return err
			}

			chs = append(chs, ch)
		}

		return nil
	})

	if err != nil {
		return []*Channel{}, err
	}

	return chs, nil
}

func (b *boltStore) FindChannelByName(name string) (*Channel, bool) {
	c := &Channel{}
	found := false
	b.db.View(func(tx *bolt.Tx) error {
		buc := tx.Bucket([]byte("channels"))
		if buc == nil {
			return errors.New("bucket does not exist")
		}
		v := buc.Get([]byte(name))
		if v != nil {
			if err := json.Unmarshal(v, c); err != nil {
				return err
			}
			found = true
			return nil
		}
		return nil
	})

	return c, found
}

func (b *boltStore) UpdateChannel(c *Channel) error {
	bs, err := json.Marshal(c)
	if err != nil {
		return err
	}

	err = b.db.Update(func(tx *bolt.Tx) error {
		buc := tx.Bucket([]byte("channels"))
		if buc == nil {
			return errors.New("bucket does not exist")
		}

		v := buc.Get([]byte(c.Name))
		if v == nil {
			return errors.New("channel " + c.Name + " doesn't exist")
		}

		err = buc.Put([]byte(c.Name), bs)
		return err
	})

	return err
}

func (b *boltStore) InsertChannel(c *Channel) error {
	bs, err := json.Marshal(c)
	if err != nil {
		return err
	}

	err = b.db.Update(func(tx *bolt.Tx) error {
		buc := tx.Bucket([]byte("channels"))
		if buc == nil {
			return errors.New("bucket does not exist")
		}

		res := buc.Get([]byte(c.Name))
		if res != nil {
			return errors.New("duplicate key error")
		}

		err = buc.Put([]byte(c.Name), bs)
		return err
	})

	return err
}

func (b *boltStore) DeleteChannel(c *Channel) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		buc := tx.Bucket([]byte("channels"))
		if buc == nil {
			return errors.New("bucket does not exist")
		}
		err := buc.Delete([]byte(c.Name))
		return err
	})
	return err
}

func (b *boltStore) Open() error {
	db, err := bolt.Open(viper.GetString("persistence.db_file"), 0600, nil)
	if err != nil {
		return err
	}
	b.db = db

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("channels"))
		return err
	})
	return err
}

func InitializeMetaStore() error {
	switch viper.GetString("persistence.store") {
	case "bolt":
		return initBoltStore()
	default:
		return initBoltStore()
	}
}

func CloseMetaStore() error {
	return MStore.Close()
}

func initBoltStore() error {
	b := &boltStore{}
	err := b.Open()
	if err != nil {
		return err
	}
	MStore = b
	return nil
}
