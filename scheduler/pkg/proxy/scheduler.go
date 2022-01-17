package proxy

import (
	"github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	"github.com/sirupsen/logrus"
)

type ModelEvent struct {
	*agent.ModelOperationMessage
}

type Scheduler struct {
	logger      logrus.FieldLogger
	modelEvents <-chan ModelEvent
	agentServer *AgentServer
}

func NewScheduler(l logrus.FieldLogger, es <-chan ModelEvent, agentServer *AgentServer) *Scheduler {
	return &Scheduler{
		logger:      l,
		modelEvents: es,
		agentServer: agentServer,
	}
}

func (s *Scheduler) Start() {
	for e := range s.modelEvents {
		opName, ok := agent.ModelOperationMessage_Operation_name[int32(e.Operation)]
		if !ok {
			opName = "unknown operation"
		}

		s.logger.Debugf(
			"received %s event for model %s",
			opName,
			e.GetModelVersion().GetModel().GetMeta().GetName(),
		)

		err := s.agentServer.HandleModelEvent(e)
		if err != nil {
			s.logger.WithError(err).Warnf("encountered issue scheduling update to model %s", e.GetModelVersion().GetModel().GetMeta().GetName())
		}
	}
}
