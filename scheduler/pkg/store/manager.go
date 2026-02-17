/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package store

import (
	"context"
	"fmt"
	"sync"

	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/utils"
	log "github.com/sirupsen/logrus"
)

type ModelServerStore interface {
	GetModel(ctx context.Context, key string) (*pb.ModelSnapshot, error)
	PutModel(ctx context.Context, key string, value *pb.ModelSnapshot) error
}

type ServerStore interface {
	GetServer(ctx context.Context)
}

type manager struct {
	mu       sync.RWMutex
	storage  ModelServerStore
	logger   log.FieldLogger
	eventHub *coordinator.EventHub
}

type ModelServerManager interface {
	UpdateModel(ctx context.Context, req *pb.LoadModelRequest) error
}

func (m *manager) UpdateModel(ctx context.Context, req *pb.LoadModelRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	modelName := req.GetModel().GetMeta().GetName()
	validName := utils.CheckName(modelName)
	if !validName {
		return fmt.Errorf(
			"Model %s does not have a valid name - it must be alphanumeric and not contains dots (.)",
			modelName,
		)
	}

	modelSnap, err := m.storage.GetModel(ctx, modelName)
	if err != nil {
		return fmt.Errorf("could not update model %s: %v", modelName, err)
	}

	if modelSnap == nil {
		err = m.storage.PutModel(ctx, modelName, NewModelSnapshot(req.GetModel()))
		if err != nil {
			return fmt.Errorf("could not create a new model %s: %v", modelName, err)
		}
		return nil
	}

	if modelSnap.GetDeleted() {
		if ModelInactive(modelSnap) {
			return fmt.Errorf("model %s is in process of deletion - new model can not be created", modelName)
		}

		modelSnap = CreateNextModelVersion(modelSnap, req.GetModel())

	}

	// todo: do Model EqualityCheck

	err = m.storage.PutModel(ctx, modelName, NewModelSnapshot(req.GetModel()))
	if err != nil {
		return fmt.Errorf("could not create a new model %s: %v", modelName, err)
	}
	return nil

}
