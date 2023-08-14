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

package modelscaling

import (
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
	log "github.com/sirupsen/logrus"
)

type DataPlaneStatsCollector struct {
	logger      log.FieldLogger
	StatKeepers []interfaces.ModelStatsKeeper
}

func NewDataPlaneStatsCollector(statKeepers []interfaces.ModelStatsKeeper, logger log.FieldLogger) *DataPlaneStatsCollector {
	return &DataPlaneStatsCollector{
		logger:      logger,
		StatKeepers: statKeepers,
	}
}

func (c *DataPlaneStatsCollector) ModelInferEnter(internalModelName, requestId string) error {
	var err error
	c.logger.Infof("ModelInferEnter for model %s request %s", internalModelName, requestId)
	for _, stat := range c.StatKeepers {
		err = stat.ModelInferEnter(internalModelName, requestId)
		if err != nil {
			c.logger.WithError(err).Warnf("model stats error for model %s request %s", internalModelName, requestId)
		}
	}

	return nil
}

func (c *DataPlaneStatsCollector) ModelInferExit(internalModelName, requestId string) error {
	var err error
	c.logger.Infof("ModelInferExit for model %s request %s", internalModelName, requestId)
	for _, stat := range c.StatKeepers {
		err = stat.ModelInferExit(internalModelName, requestId)
		if err != nil {
			c.logger.WithError(err).Warnf("model stats error for model %s request %s", internalModelName, requestId)
		}

	}
	return nil
}
