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

const (
	defaultExperimentSnapshotVersion = "v1"
	currentExperimentSnapshotVersion = "v2"
	versionKey                       = "__version_key__"
)

type ExperimentDBManager struct {
	db     *badger.DB
	logger logrus.FieldLogger
}

func newExperimentDbManager(path string, logger logrus.FieldLogger) (*ExperimentDBManager, error) {
	options := badger.DefaultOptions(path)
	options.Logger = logger.WithField("source", "experimentDb")
	db, err := badger.Open(options)
	if err != nil {
		return nil, err
	}
	return &ExperimentDBManager{
		db:     db,
		logger: logger,
	}, nil
}

func (edb *ExperimentDBManager) Stop() error {
	return edb.db.Close()
}

func (edb *ExperimentDBManager) save(experiment *Experiment) error {
	experimentProto := CreateExperimentSnapshotProto(experiment)
	experimentBytes, err := proto.Marshal(experimentProto)
	if err != nil {
		return err
	}
	return edb.db.Update(func(txn *badger.Txn) error {
		err = txn.Set([]byte(experiment.Name), experimentBytes)
		return err
	})
}

func (edb *ExperimentDBManager) saveVersion() error {
	return edb.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(versionKey), []byte(currentExperimentSnapshotVersion))
	})
}

func (edb *ExperimentDBManager) getVersion() (string, error) {
	version := defaultExperimentSnapshotVersion // defaullt version
	err := edb.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(([]byte(versionKey)))
		if err != nil {
			return err
		}
		return item.Value(func(v []byte) error {
			version = string(v)
			return nil
		})
	})
	return version, err
}

// TODO: as with pipeline deletion, we should also delete the experiment from the db once we guarantee that
// the event has been consumed by all relevant subscribers (e.g. controller, etc.)
// currently we want to replay all events on reconnection
func (edb *ExperimentDBManager) delete(name string) error {
	return edb.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(name))
		return err
	})
}

func (edb *ExperimentDBManager) restore(
	startExperimentCb func(*Experiment) error, stopExperimentCb func(*Experiment) error) error {
	return edb.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := string(item.Key())
			if key == versionKey {
				// skip the version key
				continue
			}
			err := item.Value(func(v []byte) error {
				snapshot := scheduler.ExperimentSnapshot{}
				err := proto.Unmarshal(v, &snapshot)
				if err != nil {
					return err
				}
				experiment := CreateExperimentFromSnapshot(&snapshot)
				if experiment.Deleted {
					err = stopExperimentCb(experiment)
				} else {
					// otherwise attempt to start the experiment
					err = startExperimentCb(experiment)
				}
				if err != nil {
					// If the callback fails, do not bubble the error up but simply log it as a warning.
					// The experiment restore is skipped instead of returning an error which would cause the scheduler to fail.
					edb.logger.WithError(err).Warnf("failed to restore experiment %s", experiment.Name)
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

// get experiment by name from db
func (edb *ExperimentDBManager) get(name string) (*Experiment, error) {
	var experiment *Experiment
	err := edb.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(([]byte(name)))
		if err != nil {
			return err
		}
		return item.Value(func(v []byte) error {

			snapshot := scheduler.ExperimentSnapshot{}
			err = proto.Unmarshal(v, &snapshot)
			if err != nil {
				return err
			}
			experiment = CreateExperimentFromSnapshot(&snapshot)
			return err
		})
	})
	return experiment, err
}

// migrateToDBV2 deletes all experiments from the db
// the reason why we went ahead with this approach is that the experiment that we store in the old
// format doesnt have a delete field, so we cannot distinguish between deleted and active experiments
// we then will rely on the operator to re-create the experiments from the etcd snapshot
func (edb *ExperimentDBManager) migrateToDBV2() error {
	err := edb.db.DropAll()
	if err != nil {
		return err
	}
	return edb.saveVersion()
}
