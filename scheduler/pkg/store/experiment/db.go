/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package experiment

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
)

type ExperimentDBManager struct {
	db *badger.DB
}

func newExperimentDbManager(path string, logger logrus.FieldLogger) (*ExperimentDBManager, error) {
	options := badger.DefaultOptions(path)
	options.Logger = logger.WithField("source", "experimentDb")
	db, err := badger.Open(options)
	if err != nil {
		return nil, err
	}
	return &ExperimentDBManager{
		db: db,
	}, nil
}

func (edb *ExperimentDBManager) Stop() error {
	return edb.db.Close()
}

func (edb *ExperimentDBManager) save(experiment *Experiment) error {
	experimentProto := CreateExperimentProto(experiment)
	experimentBytes, err := proto.Marshal(experimentProto)
	if err != nil {
		return err
	}
	return edb.db.Update(func(txn *badger.Txn) error {
		err = txn.Set([]byte(experiment.Name), experimentBytes)
		return err
	})
}

// TODO: as with pipeline deletion, we should also delete the experiment from the db once we guarantee that
// the event has been consumed by all relevant subscribers (e.g. controller, etc.)
// currently we want to replay all events on reconnection
// func (edb *ExperimentDBManager) delete(experiment *Experiment) error {
// 	return edb.db.Update(func(txn *badger.Txn) error {
// 		err := txn.Delete([]byte(experiment.Name))
// 		return err
// 	})
// }

func (edb *ExperimentDBManager) restore(startExperimentCb func(*Experiment) error) error {
	return edb.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(v []byte) error {
				snapshot := scheduler.Experiment{}
				err := proto.Unmarshal(v, &snapshot)
				if err != nil {
					return err
				}
				experiment := CreateExperimentFromRequest(&snapshot)
				err = startExperimentCb(experiment)
				if err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}
