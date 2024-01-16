/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package cleaner

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
)

type ModelVersionCleaner interface {
	RunCleanup(modelName string)
}

type VersionCleaner struct {
	store  store.ModelStore
	logger log.FieldLogger
}

func NewVersionCleaner(
	schedStore store.ModelStore,
	logger log.FieldLogger,
) *VersionCleaner {
	return &VersionCleaner{
		store:  schedStore,
		logger: logger.WithField("source", "VersionCleaner"),
	}
}

func (v *VersionCleaner) RunCleanup(modelName string) {
	logger := v.logger.WithField("func", "RunCleanup")
	go func() {
		err := v.cleanupOldVersions(modelName)
		if err != nil {
			logger.WithError(err).Warnf("Failed to run cleanup old versions for model %s", modelName)
		}
	}()
}

func (v *VersionCleaner) cleanupOldVersions(modelName string) error {
	logger := v.logger.WithField("func", "cleanupOldVersions")
	logger.Debugf("Cleanup model %s", modelName)

	v.store.LockModel(modelName)
	defer v.store.UnlockModel(modelName)

	// Get Model
	model, err := v.store.GetModel(modelName)
	if err != nil {
		return err
	}
	if model == nil {
		return fmt.Errorf("Can't find model with key %s", modelName)
	}
	latest := model.GetLatest()
	if latest == nil {
		return fmt.Errorf("Failed to find latest model for %s", modelName)
	}
	if latest.ModelState().State == store.ModelAvailable {
		for _, mv := range model.GetVersionsBeforeLastAvailable() {
			_, err := v.store.UnloadVersionModels(modelName, mv.GetVersion())
			if err != nil {
				return err
			}
		}
	}
	return nil
}
