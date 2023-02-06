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
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/cenkalti/backoff/v4"
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
