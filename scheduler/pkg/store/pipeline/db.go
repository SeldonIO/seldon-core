package pipeline

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	"google.golang.org/protobuf/proto"
)

type PipelineDB struct {
	db *badger.DB
}

func NewPipelineDb(path string) (*PipelineDB, error) {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, err
	}
	return &PipelineDB{
		db: db,
	}, nil
}

func (pdb *PipelineDB) Stop() error {
	return pdb.db.Close()
}

func (pdb *PipelineDB) save(pipeline *Pipeline) error {
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

func (pdb *PipelineDB) delete(pipeline *Pipeline) error {
	return pdb.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(pipeline.Name))
		return err
	})
}

func (pdb *PipelineDB) restore(store *PipelineStore) error {
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
				store.restorePipeline(pipeline)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}
