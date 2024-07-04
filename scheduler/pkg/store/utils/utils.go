package utils

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/sirupsen/logrus"
)

const (
	VersionKey = "__version_key__"
)

func Open(path string, logger logrus.FieldLogger, source string) (*badger.DB, error) {
	options := badger.DefaultOptions(path)
	options.Logger = logger.WithField("source", source)
	db, err := badger.Open(options)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func Delete(db *badger.DB, key string) error {
	return db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

func Stop(db *badger.DB) error {
	return db.Close()
}

func SaveVersion(db *badger.DB, version string) error {
	return db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(VersionKey), []byte(version))
	})
}

func GetVersion(db *badger.DB, defaultVal string) (string, error) {
	var version string = defaultVal
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(VersionKey))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			version = string(val)
			return nil
		})
	})
	return version, err
}
