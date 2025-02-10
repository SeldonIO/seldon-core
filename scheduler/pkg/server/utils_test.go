/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package server

import (
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
)

func TestSendWithTimeout(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name      string
		sleepTime time.Duration
		err       error
		isErr     bool
		isExpired bool
	}

	fn := func(err error) error {
		time.Sleep(5 * time.Millisecond)
		return err
	}

	tests := []test{
		{
			name:      "simple",
			sleepTime: 10 * time.Millisecond,
			err:       nil,
			isErr:     false,
			isExpired: false,
		},
		{
			name:      "timeout",
			sleepTime: 1 * time.Millisecond,
			err:       nil,
			isErr:     true,
			isExpired: true,
		},
		{
			name:      "error",
			sleepTime: 10 * time.Millisecond,
			err:       fmt.Errorf("error"),
			isErr:     true,
			isExpired: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			hasExpired, err := sendWithTimeout(func() error {
				return fn(test.err)
			}, test.sleepTime)
			g.Expect(hasExpired).To(Equal(test.isExpired))
			if test.isErr {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
			}
		})
	}
}

func TestShouldScaleDown(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name            string
		server          *store.ServerSnapshot
		shouldScaleDown bool
	}

	tests := []test{
		{
			name: "should scale down - empty replicas",
			server: &store.ServerSnapshot{
				Stats: &store.ServerStats{
					NumEmptyReplicas:          1,
					MaxNumReplicaHostedModels: 0,
				},
				ExpectedReplicas: 2,
			},
			shouldScaleDown: true,
		},
		{
			name: "should scale down - pack",
			server: &store.ServerSnapshot{
				Stats: &store.ServerStats{
					NumEmptyReplicas:          0,
					MaxNumReplicaHostedModels: 1,
				},
				ExpectedReplicas: 2,
			},
			shouldScaleDown: true,
		},
		{
			name: "should not scale down - empty replicas - last replica",
			server: &store.ServerSnapshot{
				Stats: &store.ServerStats{
					NumEmptyReplicas:          1,
					MaxNumReplicaHostedModels: 0,
				},
				ExpectedReplicas: 1,
			},
			shouldScaleDown: false,
		},
		{
			name: "should not scale down - pack - last replica",
			server: &store.ServerSnapshot{
				Stats: &store.ServerStats{
					NumEmptyReplicas:          1,
					MaxNumReplicaHostedModels: 0,
				},
				ExpectedReplicas: 1,
			},
			shouldScaleDown: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g.Expect(shouldScaleDown(test.server, 1.0)).To(Equal(test.shouldScaleDown))
		})
	}
}
