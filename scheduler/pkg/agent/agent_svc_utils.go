/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	boff "github.com/cenkalti/backoff/v4"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/config"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
)

func logSubserviceNotYetReady(logger *log.Entry, subserviceName string) boff.Notify {
	return func(err error, delay time.Duration) {
		logger.WithError(err).Infof("Waiting for %s service to become ready... next check in %s", subserviceName, delay.Round(100*time.Millisecond))
	}
}

func newAgentServiceStatusRetryBackoff(ctx context.Context) boff.BackOff {
	expBackoff := boff.NewExponentialBackOff(boff.WithMaxInterval(config.ServiceReadyRetryMaxInterval))
	return boff.WithContext(expBackoff, ctx)
}

func startSubService(
	service interfaces.DependencyServiceInterface,
	logger *log.Entry,
	ctx context.Context,
) error {
	logger.Infof("Starting and waiting for %s", service.Name())
	err := service.Start()
	if err != nil {
		return err
	}

	return isReady(service, logger, ctx)
}

func isReady(service interfaces.DependencyServiceInterface, logger *log.Entry, ctx context.Context) error {
	backoffWithContext := newAgentServiceStatusRetryBackoff(ctx)

	logFailure := logSubserviceNotYetReady(logger, service.Name())

	readyToError := func() error {
		if service.Ready() {
			return nil
		} else {
			return fmt.Errorf("service %s not ready", service.Name())
		}
	}
	return boff.RetryNotify(readyToError, backoffWithContext, logFailure)
}

func getModifiedModelVersion(modelId string, version uint32, originalModelVersion *agent.ModelVersion, modelRuntimeInfo *scheduler.ModelRuntimeInfo) *agent.ModelVersion {
	mv := proto.Clone(originalModelVersion)
	mv.(*agent.ModelVersion).Model.Meta.Name = modelId
	if modelRuntimeInfo != nil && modelRuntimeInfo.ModelRuntimeInfo != nil {
		if mv.(*agent.ModelVersion).Model.ModelSpec == nil {
			mv.(*agent.ModelVersion).Model.ModelSpec = &scheduler.ModelSpec{}
		}
		mv.(*agent.ModelVersion).Model.ModelSpec.ModelRuntimeInfo = modelRuntimeInfo
	}
	mv.(*agent.ModelVersion).Version = version
	return mv.(*agent.ModelVersion)
}

func isReadyChecker(
	isStartup bool,
	service interfaces.DependencyServiceInterface,
	logger *log.Entry,
	readyNotifications chan<- SubServiceReadinessNotification,
	ctx context.Context,
) {
	var err error
	if isStartup {
		err = startSubService(service, logger, ctx)
	} else {
		err = isReady(service, logger, ctx)
	}

	readyNotifications <- SubServiceReadinessNotification{
		err:            err,
		subserviceName: service.Name(),
		subServiceType: service.GetType(),
	}
}

func backoffWithMaxNumRetry(fn func() error, count uint8, maxElapsedTime time.Duration, logger log.FieldLogger) error {
	backoffWithMax := boff.NewExponentialBackOff(boff.WithMaxElapsedTime(maxElapsedTime))
	i := 0
	logFailure := func(err error, delay time.Duration) {
		logger.WithError(err).Errorf("Retry op #%d", i)
		i++
	}
	return boff.RetryNotify(fn, newBackOffWithMaxCount(count, backoffWithMax), logFailure)
}

// backOffWithMaxCount is a backoff policy that retries up to a max count
type backOffWithMaxCount struct {
	backoffPolicy boff.BackOff
	maxCount      uint8
	currentCount  uint8
}

func newBackOffWithMaxCount(maxCount uint8, backOffPolicy boff.BackOff) *backOffWithMaxCount {
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
	if b.currentCount >= b.maxCount-1 {
		return boff.Stop
	} else {
		b.currentCount++
		return b.backoffPolicy.NextBackOff()
	}
}

func ignoreIfOutOfOrder(key string, timestamp int64, timestamps *sync.Map) bool {
	tick, ok := timestamps.Load(key)
	if !ok {
		timestamps.Store(key, timestamp)
	} else {
		if timestamp < tick.(int64) {
			return true
		}
	}
	return false
}
