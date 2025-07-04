/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package pipeline

import (
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/utils"
)

const (
	defaultPipelineSnapshotVersion = "v1"
	currentPipelineSnapshotVersion = "v2"
)

type PipelineDBManager struct {
	db                 *badger.DB
	logger             logrus.FieldLogger
	deletedResourceTTL time.Duration
}

func newPipelineDbManager(path string, logger logrus.FieldLogger, deletedResourceTTL uint) (*PipelineDBManager, error) {
	db, err := utils.Open(path, logger, "pipelineDb")
	if err != nil {
		return nil, err
	}

	pdb := &PipelineDBManager{
		db:                 db,
		logger:             logger,
		deletedResourceTTL: time.Duration(deletedResourceTTL * uint(time.Second)),
	}

	version, err := pdb.getVersion()
	if err != nil {
		// assume that if the version key is not found then either:
		// -  the db is in the old format
		// -  the db is empty
		// in either case we will migrate the db to the current version
		logger.Infof("Migrating DB from version %s to %s", version, currentPipelineSnapshotVersion)
		err := pdb.migrateToDBCurrentVersion()
		if err != nil {
			return nil, err
		}
	}
	// in the furture we can add migration logic here for > v1
	return pdb, nil
}

// a nil ttl will save the pipeline indefinitely
func (pdb *PipelineDBManager) save(pipeline *Pipeline) error {
	pipelineProto := CreatePipelineSnapshotFromPipeline(pipeline)
	pipelineBytes, err := proto.Marshal(pipelineProto)
	if err != nil {
		return err
	}
	if !pipeline.Deleted {
		return pdb.db.Update(func(txn *badger.Txn) error {
			err = txn.Set([]byte(pipeline.Name), pipelineBytes)
			return err
		})
	} else {
		return pdb.db.Update(func(txn *badger.Txn) error {
			e := badger.NewEntry([]byte(pipeline.Name), pipelineBytes).WithTTL(pdb.deletedResourceTTL)
			err = txn.SetEntry(e)
			return err
		})
	}
}

func (pdb *PipelineDBManager) Stop() error {
	return utils.Stop(pdb.db)
}

// TODO: delete unused pipelines from the store as for now it increases indefinitely
func (pdb *PipelineDBManager) delete(name string) error {
	return utils.Delete(pdb.db, name)
}

func (pdb *PipelineDBManager) saveVersion() error {
	return utils.SaveVersion(pdb.db, currentPipelineSnapshotVersion)
}

func (pdb *PipelineDBManager) getVersion() (string, error) {
	return utils.GetVersion(pdb.db, defaultPipelineSnapshotVersion)
}

func (pdb *PipelineDBManager) restore(createPipelineCb func(pipeline *Pipeline)) error {
	return pdb.db.View(func(txn *badger.Txn) error {
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
				snapshot := scheduler.PipelineSnapshot{}
				err := proto.Unmarshal(v, &snapshot)
				if err != nil {
					return err
				}

				pipeline, err := CreatePipelineFromSnapshot(&snapshot)
				if err != nil {
					return err
				}

				if pipeline.Deleted {
					pipeline.DeletedAt = utils.GetDeletedAt(item, pdb.deletedResourceTTL)
				}

				createPipelineCb(pipeline)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// get pipeline by name from db
func (pdb *PipelineDBManager) get(name string) (*Pipeline, error) {
	var pipeline *Pipeline
	err := pdb.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(([]byte(name)))
		if err != nil {
			return err
		}
		return item.Value(func(v []byte) error {
			snapshot := scheduler.PipelineSnapshot{}
			err = proto.Unmarshal(v, &snapshot)
			if err != nil {
				return err
			}
			pipeline, err = CreatePipelineFromSnapshot(&snapshot)
			if err != nil {
				return err
			}
			return err
		})
	})
	return pipeline, err
}

// we only save the version key in the db for this migration
func (pdb *PipelineDBManager) migrateToDBCurrentVersion() error {
	return pdb.saveVersion()
}
