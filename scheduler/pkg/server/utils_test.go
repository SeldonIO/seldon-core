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

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler/db"

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

func TestSouldScaleUp(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                string
		shouldScaleUp       bool
		newExpectedReplicas uint32
		server              *db.Server
		stats               *store.ServerStats
	}

	tests := []test{
		{
			name:                "scales up to MaxReplicas",
			shouldScaleUp:       true,
			newExpectedReplicas: 2,
			server: &db.Server{
				MaxReplicas:      2,
				ExpectedReplicas: 1,
			},
			stats: &store.ServerStats{MaxNumReplicaHostedModels: 3},
		},
		{
			name:                "scales up to MaxNumReplicaHostedModels",
			shouldScaleUp:       true,
			newExpectedReplicas: 3,
			server: &db.Server{
				MaxReplicas:      4,
				ExpectedReplicas: 1,
			},
			stats: &store.ServerStats{MaxNumReplicaHostedModels: 3},
		},
		{
			name:          "should not scale if expectedReplicas is greater than MaxNumReplicaHostedModels",
			shouldScaleUp: false,
			server: &db.Server{
				MaxReplicas:      3,
				ExpectedReplicas: 3,
			},
			stats: &store.ServerStats{MaxNumReplicaHostedModels: 2},
		},
		{
			name:          "does not scale up for ExpectedReplicas below 0",
			shouldScaleUp: false,
			server: &db.Server{
				MaxReplicas:      2,
				ExpectedReplicas: -1,
			},
			stats: &store.ServerStats{MaxNumReplicaHostedModels: 3},
		},
		{
			name:          "does not scale up for missing max replicas",
			shouldScaleUp: false,
			server: &db.Server{
				ExpectedReplicas: 1,
			},
			stats: &store.ServerStats{MaxNumReplicaHostedModels: 3},
		},
		{
			name:          "does not scale to zero",
			shouldScaleUp: false,
			server: &db.Server{
				MaxReplicas:      0,
				ExpectedReplicas: 0,
			},
			stats: &store.ServerStats{MaxNumReplicaHostedModels: 0},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ok, expectedReplicas := shouldScaleUp(test.server, test.stats)
			g.Expect(ok).To(Equal(test.shouldScaleUp))
			if test.shouldScaleUp {
				g.Expect(expectedReplicas).To(Equal(test.newExpectedReplicas))
			}
		})
	}
}

func TestShouldScaleDown(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name             string
		server           *db.Server
		stats            *store.ServerStats
		shouldScaleDown  bool
		expectedReplicas uint32
		packThreshold    float32
	}

	tests := []test{
		{
			name: "should scale down - empty replicas",
			server: &db.Server{
				ExpectedReplicas: 2,
				MinReplicas:      1,
			},
			stats: &store.ServerStats{
				NumEmptyReplicas:          1,
				MaxNumReplicaHostedModels: 0,
			},
			shouldScaleDown:  true,
			expectedReplicas: 1,
			packThreshold:    0.0,
		},
		{
			name: "should scale down - empty replicas > 1 - 1",
			server: &db.Server{
				ExpectedReplicas: 3,
				MinReplicas:      1,
			},
			stats: &store.ServerStats{
				NumEmptyReplicas:          2,
				MaxNumReplicaHostedModels: 0,
			},
			shouldScaleDown:  true,
			expectedReplicas: 1,
			packThreshold:    0.0,
		},
		{
			name: "should scale down - violate min replicas",
			server: &db.Server{
				ExpectedReplicas: 3,
				MinReplicas:      2,
			},
			stats: &store.ServerStats{
				NumEmptyReplicas:          2,
				MaxNumReplicaHostedModels: 0,
			},
			shouldScaleDown:  true,
			expectedReplicas: 2,
			packThreshold:    0.0,
		},
		{
			name: "should scale down - empty replicas > 1 - 2",
			server: &db.Server{
				ExpectedReplicas: 3,
				MinReplicas:      1,
			},
			stats: &store.ServerStats{
				NumEmptyReplicas:          1,
				MaxNumReplicaHostedModels: 0,
			},
			shouldScaleDown:  true,
			expectedReplicas: 2,
			packThreshold:    0.0,
		},
		{
			name: "should scale down - pack",
			server: &db.Server{
				ExpectedReplicas: 2,
				MinReplicas:      1,
			},
			stats: &store.ServerStats{
				NumEmptyReplicas:          0,
				MaxNumReplicaHostedModels: 1,
			},
			shouldScaleDown:  true,
			expectedReplicas: 1,
			packThreshold:    1.0,
		},
		{
			name: "should scale down - pack > 1",
			server: &db.Server{
				ExpectedReplicas: 3,
				MinReplicas:      1,
			},
			stats: &store.ServerStats{
				NumEmptyReplicas:          0,
				MaxNumReplicaHostedModels: 1,
			},
			shouldScaleDown:  true,
			expectedReplicas: 1,
			packThreshold:    1.0,
		},
		{
			name: "should not scale down - pack threshold",
			server: &db.Server{
				ExpectedReplicas: 3,
				MinReplicas:      1,
			},
			stats: &store.ServerStats{
				NumEmptyReplicas:          0,
				MaxNumReplicaHostedModels: 1,
			},
			shouldScaleDown:  false,
			expectedReplicas: 0,
			packThreshold:    0.0,
		},
		{
			name: "should not scale down - empty replicas - last replica",
			server: &db.Server{
				ExpectedReplicas: 1,
				MinReplicas:      1,
			},
			stats: &store.ServerStats{
				NumEmptyReplicas:          1,
				MaxNumReplicaHostedModels: 0,
			},
			shouldScaleDown:  false,
			expectedReplicas: 0,
			packThreshold:    0.0,
		},
		{
			name: "should not scale down - pack - last replica",
			server: &db.Server{
				ExpectedReplicas: 1,
				MinReplicas:      1,
			},
			stats: &store.ServerStats{
				NumEmptyReplicas:          1,
				MaxNumReplicaHostedModels: 0,
			},
			shouldScaleDown:  false,
			expectedReplicas: 0,
			packThreshold:    1.0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scaleDown, replicas := shouldScaleDown(test.server, test.stats, test.packThreshold)
			g.Expect(scaleDown).To(Equal(test.shouldScaleDown))
			if scaleDown {
				g.Expect(replicas).To(Equal(test.expectedReplicas))
			}
		})
	}
}
