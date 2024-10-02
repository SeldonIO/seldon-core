/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package experiment

import (
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/utils"
)

const (
	defaultExperimentSnapshotVersion = "v1"
	currentExperimentSnapshotVersion = "v2"
)

type ExperimentDBManager struct {
	db     *badger.DB
	logger logrus.FieldLogger
}

func newExperimentDbManager(path string, logger logrus.FieldLogger) (*ExperimentDBManager, error) {
	db, err := utils.Open(path, logger, "experimentDb")
	if err != nil {
		return nil, err
	}

	edb := &ExperimentDBManager{
		db:     db,
		logger: logger,
	}

	version, err := edb.getVersion()
	if err != nil {
		// assume that if the version key is not found then either:
		// -  the db is in the old format
		// -  the db is empty
		// in either case we will migrate the db to the current version
		logger.Infof("Migrating DB from version %s to %s", version, currentExperimentSnapshotVersion)
		err := edb.migrateToDBCurrentVersion()
		if err != nil {
			return nil, err
		}
	}
	// in the furture we can add migration logic here for > v1
	return edb, nil
}

func (edb *ExperimentDBManager) Stop() error {
	return utils.Stop(edb.db)
}

// a nil ttl will save the experiment indefinitely
func (edb *ExperimentDBManager) save(experiment *Experiment, ttl *time.Duration) error {
	experimentProto := CreateExperimentSnapshotProto(experiment)
	experimentBytes, err := proto.Marshal(experimentProto)
	if err != nil {
		return err
	}

	if ttl == nil {
		return edb.db.Update(func(txn *badger.Txn) error {
			err = txn.Set([]byte(experiment.Name), experimentBytes)
			return err
		})
	} else {
		return edb.db.Update(func(txn *badger.Txn) error {
			e := badger.NewEntry([]byte(experiment.Name), experimentBytes).WithTTL(*ttl)
			err = txn.SetEntry(e)
			return err
		})
	}
}

func (edb *ExperimentDBManager) saveVersion() error {
	return utils.SaveVersion(edb.db, currentExperimentSnapshotVersion)
}

func (edb *ExperimentDBManager) getVersion() (string, error) {
	return utils.GetVersion(edb.db, defaultExperimentSnapshotVersion)
}

// TODO: as with pipeline deletion, we should also delete the experiment from the db once we guarantee that
// the event has been consumed by all relevant subscribers (e.g. controller, etc.)
// currently we want to replay all events on reconnection
func (edb *ExperimentDBManager) delete(name string) error {
	return utils.Delete(edb.db, name)
}

func (edb *ExperimentDBManager) restore(
	startExperimentCb func(*Experiment) error, stopExperimentCb func(*Experiment) error,
) error {
	return edb.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := string(item.Key())
			if key == utils.VersionKey {
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
				if experiment.Deleted && item.ExpiresAt() == 0 {
					ttl := deletedExperimentTTL
					err = edb.save(experiment, &ttl)
					if err != nil {
						edb.logger.WithError(err).Warnf("failed to set ttl for experiment %s", experiment.Name)
					}
				}
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

// migrateToDBCurrentVersion deletes all experiments from the db
// the reason why we went ahead with this approach is that the experiment that we store in the old
// format doesnt have a delete field, so we cannot distinguish between deleted and active experiments
// we then will rely on the operator to re-create the experiments from the etcd snapshot
func (edb *ExperimentDBManager) migrateToDBCurrentVersion() error {
	err := edb.db.DropAll()
	if err != nil {
		return err
	}
	return edb.saveVersion()
}
