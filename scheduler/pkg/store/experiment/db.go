package experiment

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
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

func (edb *ExperimentDBManager) delete(experiment *Experiment) error {
	return edb.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(experiment.Name))
		return err
	})
}

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
				if err != nil {
					return err
				}
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
