package kv

import "github.com/dgraph-io/badger/v2"

type Database struct {
	db *badger.DB
}

func NewDatabase(path string) (*Database, error) {
	opts := badger.DefaultOptions(path)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) MultiPut(keys [][]byte, values [][]byte) error {
	return d.db.Update(func(txn *badger.Txn) error {
		var err error
		for i := 0; i < len(keys); i++ {
			if err = txn.Set(keys[i], values[i]); err != nil {
				return err
			}
		}
		return nil
	})
}

func (d *Database) Put(key []byte, value []byte) error {
	return d.db.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, value)
		return err
	})
}

func (d *Database) Delete(key []byte) error {
	return d.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete(key)
		return err
	})
}

func (d *Database) Get(key []byte) ([]byte, error) {
	var result []byte
	err := d.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			result = append([]byte{}, val...)
			return nil
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (d *Database) Has(key []byte) (bool, error) {
	has := false
	err := d.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		if err != nil {
			if err != badger.ErrKeyNotFound {
				return err
			}
			return nil
		}
		has = true
		return nil
	})
	return has, err
}

func (d *Database) Update(key []byte, updateFunc func(value []byte) ([]byte, error)) error {
	return d.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		var oldValue []byte
		err = item.Value(func(val []byte) error {
			oldValue = append([]byte{}, val...)
			return nil
		})
		if err != nil {
			return err
		}
		newValue, err := updateFunc(oldValue)
		if err != nil {
			return err
		}
		return txn.Set(key, newValue)
	})
}
