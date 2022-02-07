package cleaner

import (
	"fmt"

	"github.com/seldonio/seldon-core/scheduler/pkg/coordinator"

	"github.com/seldonio/seldon-core/scheduler/pkg/store"
	log "github.com/sirupsen/logrus"
)

type VersionCleaner struct {
	store     store.SchedulerStore
	logger    log.FieldLogger
	chanEvent chan coordinator.ModelEventMsg
}

func NewVersionCleaner(schedStore store.SchedulerStore, logger log.FieldLogger, eventHub *coordinator.ModelEventHub) *VersionCleaner {
	v := &VersionCleaner{
		store:     schedStore,
		logger:    logger.WithField("source", "VersionCleaner"),
		chanEvent: make(chan coordinator.ModelEventMsg, 1),
	}
	eventHub.AddListener(v.chanEvent)
	return v
}

func (v *VersionCleaner) ListenForEvents() {
	logger := v.logger.WithField("func", "ListenForEvents")
	for evt := range v.chanEvent {
		logger.Infof("Got model state change for %s", evt.String())
		modelEventMsg := evt
		go func() {
			err := v.cleanupOldVersions(modelEventMsg.ModelName)
			if err != nil {
				logger.WithError(err).Warnf("Failed to run cleanup old versions for model %s", modelEventMsg.String())
			}
		}()
	}
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
