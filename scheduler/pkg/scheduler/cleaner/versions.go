package cleaner

import (
	"fmt"

	"github.com/seldonio/seldon-core/scheduler/pkg/agent"
	"github.com/seldonio/seldon-core/scheduler/pkg/store"
	log "github.com/sirupsen/logrus"
)

type VersionCleaner struct {
	store  store.SchedulerStore
	logger log.FieldLogger
	events chan *store.ModelSnapshot
	agent  agent.AgentHandler
}

func NewVersionCleaner(schedStore store.SchedulerStore, logger log.FieldLogger, agent agent.AgentHandler) *VersionCleaner {
	v := &VersionCleaner{
		store:  schedStore,
		logger: logger,
		events: make(chan *store.ModelSnapshot, 1),
		agent:  agent,
	}
	schedStore.AddListener(v.events) // Add ourselves to listen for status updates
	return v
}

func (v *VersionCleaner) ListenForEvents() {
	logger := v.logger.WithField("func", "ListenForEvents")
	for modelSnapshot := range v.events {
		logger.Infof("Got model state change for %s", modelSnapshot.Name)
		modelName := modelSnapshot.Name
		go func() {
			err := v.cleanupOldVersions(modelName)
			if err != nil {
				logger.WithError(err).Warnf("Failed to run cleanup old versions for model %s", modelName)
			}
		}()
	}
}

func (v *VersionCleaner) StopListenForEvents() {
	close(v.events)
}

func (v *VersionCleaner) cleanupOldVersions(modelName string) error {
	logger := v.logger.WithField("func", "cleanupOldVersions")
	logger.Debugf("Schedule model %s", modelName)
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
			err = v.store.UpdateLoadedModels(modelName, mv.GetVersion(), mv.Server(), []*store.ServerReplica{})
			if err != nil {
				return err
			}
		}
	}
	v.agent.SendAgentSync(modelName)
	return nil
}
