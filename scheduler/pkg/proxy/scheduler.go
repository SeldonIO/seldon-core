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

package proxy

import (
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"
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
