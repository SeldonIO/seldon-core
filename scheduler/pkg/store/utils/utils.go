/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package utils

import (
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/sirupsen/logrus"
)

const (
	VersionKey                       = "__version_key__"
	DeletedResourceTTL time.Duration = time.Duration(24 * time.Hour)
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

func GetDeletesAt(item *badger.Item) time.Time {
	return time.Unix(int64(item.ExpiresAt()), 0).Add(DeletedResourceTTL)
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
