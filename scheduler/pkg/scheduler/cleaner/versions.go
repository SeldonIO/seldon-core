package cleaner

import (
	"fmt"

	"github.com/seldonio/seldon-core/scheduler/pkg/coordinator"

	"github.com/seldonio/seldon-core/scheduler/pkg/store"
	log "github.com/sirupsen/logrus"
)

const (
	pendingEventsQueueSize int = 100
	modelEventHandlerName      = "version.cleaner.models"
)

type VersionCleaner struct {
	store  store.ModelStore
	logger log.FieldLogger
}

func NewVersionCleaner(
	schedStore store.ModelStore,
	logger log.FieldLogger,
	eventHub *coordinator.EventHub,
) *VersionCleaner {
	v := &VersionCleaner{
		store:  schedStore,
		logger: logger.WithField("source", "VersionCleaner"),
	}

	eventHub.RegisterModelEventHandler(
		modelEventHandlerName,
		pendingEventsQueueSize,
		v.logger,
		v.handleEvents,
	)

	return v
}

func (v *VersionCleaner) handleEvents(event coordinator.ModelEventMsg) {
	logger := v.logger.WithField("func", "ListenForEvents")
	logger.Infof("Got model state change for %s", event.String())

	go func() {
		err := v.cleanupOldVersions(event.ModelName)
		if err != nil {
			logger.WithError(err).Warnf("Failed to run cleanup old versions for model %s", event.String())
		}
	}()
}

func (v *VersionCleaner) cleanupOldVersions(modelName string) error {
	logger := v.logger.WithField("func", "cleanupOldVersions")
	logger.Debugf("Schedule model %s", modelName)

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
