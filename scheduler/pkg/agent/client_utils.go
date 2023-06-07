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

package agent

import (
	"fmt"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
)

func startSubService(
	service interfaces.DependencyServiceInterface,
	logger *log.Entry,
	maxElapsedTimeReadySubServiceBeforeStart time.Duration,
) error {
	logger.Infof("Starting and waiting for %s", service.Name())
	err := service.Start()
	if err != nil {
		return err
	}

	return isReady(service, logger, maxElapsedTimeReadySubServiceBeforeStart)
}

func isReady(service interfaces.DependencyServiceInterface, logger *log.Entry, maxElapsedTime time.Duration) error {
	logFailure := func(err error, delay time.Duration) {
		logger.WithError(err).Errorf("%s service not ready", service.Name())
	}

	readyToError := func() error {
		if service.Ready() {
			return nil
		} else {
			return fmt.Errorf("Service %s not ready", service.Name())
		}
	}
	backoffWithMax := backoff.NewExponentialBackOff()
	backoffWithMax.MaxElapsedTime = maxElapsedTime
	return backoff.RetryNotify(readyToError, backoffWithMax, logFailure)
}

func getModifiedModelVersion(modelId string, version uint32, originalModelVersion *agent.ModelVersion) *agent.ModelVersion {
	mv := proto.Clone(originalModelVersion)
	mv.(*agent.ModelVersion).Model.Meta.Name = modelId
	mv.(*agent.ModelVersion).Version = version
	return mv.(*agent.ModelVersion)
}

func isReadyChecker(
	isStartup bool,
	service interfaces.DependencyServiceInterface,
	logger *log.Entry,
	logMessage string,
	maxElapsedTime time.Duration,
) error {
	if isStartup {
		if err := startSubService(service, logger, maxElapsedTime); err != nil {
			logger.WithError(err).Error(logMessage)
			return err
		}
	} else {
		if err := isReady(service, logger, maxElapsedTime); err != nil {
			logger.WithError(err).Error(logMessage + " - after agent start")
			return err
		}
	}
	return nil
}

func backoffWithMaxNumRetry(fn func() error, count uint8, logger log.FieldLogger) error {
	backoffWithMax := backoff.NewExponentialBackOff()
	// Wait for model repo to be ready
	i := 0
	logFailure := func(err error, delay time.Duration) {
		logger.WithError(err).Errorf("Retry op #%d", i)
		i++
	}
	return backoff.RetryNotify(fn, newBackOffWithMaxCount(count, backoffWithMax), logFailure)
}

// backOffWithMaxCount is a backoff policy that retries up to a max count
type backOffWithMaxCount struct {
	backoffPolicy backoff.BackOff
	maxCount      uint8
	currentCount  uint8
}

func newBackOffWithMaxCount(maxCount uint8, backOffPolicy backoff.BackOff) *backOffWithMaxCount {
	return &backOffWithMaxCount{
		maxCount:      maxCount,
		backoffPolicy: backOffPolicy,
		currentCount:  0,
	}
}

func (b *backOffWithMaxCount) Reset() {
	b.backoffPolicy.Reset()
}

func (b *backOffWithMaxCount) NextBackOff() time.Duration {
	if b.currentCount >= b.maxCount {
		return backoff.Stop
	} else {
		b.currentCount++
		return b.backoffPolicy.NextBackOff()
	}
}
