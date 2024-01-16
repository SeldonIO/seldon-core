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
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

func TestBackOffPolicyWithMax(t *testing.T) {
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
			count := uint8(0)
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
				g.Expect(count).To(Equal(uint8(0)))
			}
		})
	}
}

func TestFnWrapperWithMax(t *testing.T) {
	t.Logf("Started")
	logger := log.New()
	log.SetLevel(log.DebugLevel)

	type test struct {
		name  string
		count uint8
	}
	tests := []test{
		{
			name:  "simple",
			count: 3,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			fn := func() error {
				return fmt.Errorf("error")
			}
			_ = backoffWithMaxNumRetry(fn, test.count, logger)
			// if we are here we are done
		})
	}
}
