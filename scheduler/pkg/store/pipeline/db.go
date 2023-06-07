/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
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

func (pdb *PipelineDBManager) delete(pipeline *Pipeline) error {
	return pdb.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(pipeline.Name))
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
