/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package pipeline

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
)

type PipelineDBManager struct {
	db *badger.DB
}

func newPipelineDbManager(path string, logger logrus.FieldLogger) (*PipelineDBManager, error) {
	options := badger.DefaultOptions(path)
	options.Logger = logger.WithField("source", "pipelineDb")
	db, err := badger.Open(options)
	if err != nil {
		return nil, err
	}
	return &PipelineDBManager{
		db: db,
	}, nil
}

func (pdb *PipelineDBManager) Stop() error {
	return pdb.db.Close()
}

func (pdb *PipelineDBManager) save(pipeline *Pipeline) error {
	pipelineProto := CreatePipelineSnapshotFromPipeline(pipeline)
	pipelineBytes, err := proto.Marshal(pipelineProto)
	if err != nil {
		return err
	}
	return pdb.db.Update(func(txn *badger.Txn) error {
		err = txn.Set([]byte(pipeline.Name), pipelineBytes)
		return err
	})
}

// TODO: delete unused pipelines from the store as for now it increases indefinitely
func (pdb *PipelineDBManager) delete(name string) error {
	return pdb.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(name))
		return err
	})
}

func (pdb *PipelineDBManager) restore(createPipelineCb func(pipeline *Pipeline)) error {
	return pdb.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
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

// get experiment by name from db
func (edb *PipelineDBManager) get(name string) (*Pipeline, error) {
	var pipeline *Pipeline
	err := edb.db.View(func(txn *badger.Txn) error {
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
