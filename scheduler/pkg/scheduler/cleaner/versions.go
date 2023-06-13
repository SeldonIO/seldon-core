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
