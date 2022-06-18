package pipeline

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
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
