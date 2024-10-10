/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package agent

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

func TestBackOffPolicyWithMaxCount(t *testing.T) {
	t.Logf("Started")
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	g := NewGomegaWithT(t)

	type test struct {
		name  string
		count uint8
		err   error
	}
	tests := []test{
		{
			name:  "simple",
			count: 3,
			err:   fmt.Errorf("retry"),
		},
		{
			name:  "no retry",
			count: 0,
			err:   fmt.Errorf("retry"),
		},
		{
			name:  "no error",
			count: 3,
			err:   nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			policy := backoff.ZeroBackOff{}
			fn := func() error {
				return test.err
			}
			count := uint8(1) // first call is not a retry
			policyWithMax := newBackOffWithMaxCount(test.count, &policy)
			logFailure := func(err error, delay time.Duration) {
				logger.WithError(err).Errorf("retry")
				count++
			}

			//TODO make retry configurable
			_ = backoff.RetryNotify(fn, policyWithMax, logFailure)
			if test.err != nil {
				g.Expect(count).To(Equal(test.count))
			} else {
				g.Expect(count).To(Equal(uint8(1)))
			}
		})
	}
}

func TestFnWrapperWithMax(t *testing.T) {
	t.Logf("Started")
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	g := NewGomegaWithT(t)

	type test struct {
		name           string
		count          uint8
		maxElapsedTime time.Duration
		expectedCount  uint8
	}
	tests := []test{
		{
			name:           "count > maxElapsedTime",
			count:          4,
			expectedCount:  4,
			maxElapsedTime: 30 * time.Second,
		},
		{
			name:           "count < maxElapsedTime",
			count:          4,
			expectedCount:  1,
			maxElapsedTime: 1 * time.Millisecond,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			retries := uint8(0)
			fn := func() error {
				time.Sleep(1 * time.Millisecond)
				retries++
				return fmt.Errorf("error")
			}
			_ = backoffWithMaxNumRetry(fn, test.count, test.maxElapsedTime, logger)
			// if we are here we are done
			g.Expect(retries).To(Equal(test.expectedCount))
		})
	}
}

func TestOutOfOrderUtil(t *testing.T) {
	ticks := sync.Map{}
	ticks.Store("key", time.Now().Unix())

	type test struct {
		name         string
		ticks        *sync.Map
		key          string
		timestamp    int64
		isOutOfOrder bool
	}
	tests := []test{
		{
			name:         "empty",
			ticks:        &sync.Map{},
			key:          "key",
			timestamp:    time.Now().Unix(), // dummy
			isOutOfOrder: false,
		},
		{
			name:         "in order",
			ticks:        &ticks,
			key:          "key",
			timestamp:    time.Now().Unix() + 10,
			isOutOfOrder: false,
		},
		{
			name:         "out of order",
			ticks:        &ticks,
			key:          "key",
			timestamp:    time.Now().Unix() - 10,
			isOutOfOrder: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			outOfOrder := ignoreIfOutOfOrder(test.key, test.timestamp, test.ticks)
			g.Expect(outOfOrder).To(Equal(test.isOutOfOrder))
		})
	}
}
