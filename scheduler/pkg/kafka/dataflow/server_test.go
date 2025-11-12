/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package dataflow

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/chainer"
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	kafka_config "github.com/seldonio/seldon-core/components/kafka/v2/pkg/config"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/internal/testing_utils"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka"
	cr "github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/conflict-resolution"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline/mock"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
	mock2 "github.com/seldonio/seldon-core/scheduler/v2/pkg/util/mock"
)

func TestPollerFailedTerminatingPipelines(t *testing.T) {
	tests := []struct {
		name                  string
		failedPipelines       map[string]pipeline.PipelineVersion
		needsLoadBalancer     bool
		needsConflictResolver bool
		setupMocks            func(mockPipelineHandler *mock.MockPipelineHandler, mockLoadBalancer *mock2.MockLoadBalancer, failedPipelines map[string]pipeline.PipelineVersion)
		contextTimeout        time.Duration
		tickDuration          time.Duration
		validateResult        func(g *WithT, server *ChainerServer)
		expectGomegaWithT     bool
	}{
		{
			name:                  "should return when context is cancelled",
			failedPipelines:       make(map[string]pipeline.PipelineVersion),
			needsLoadBalancer:     false,
			needsConflictResolver: false,
			setupMocks: func(mockPipelineHandler *mock.MockPipelineHandler, mockLoadBalancer *mock2.MockLoadBalancer, failedPipelines map[string]pipeline.PipelineVersion) {
				// No expectations - context cancelled before first tick
			},
			contextTimeout:    0, // Cancel immediately
			tickDuration:      100 * time.Millisecond,
			expectGomegaWithT: false,
		},
		{
			name:                  "should skip processing when no failed pipelines exist",
			failedPipelines:       make(map[string]pipeline.PipelineVersion),
			needsLoadBalancer:     false,
			needsConflictResolver: false,
			setupMocks: func(mockPipelineHandler *mock.MockPipelineHandler, mockLoadBalancer *mock2.MockLoadBalancer, failedPipelines map[string]pipeline.PipelineVersion) {
				// No expectations - empty map means no processing
			},
			contextTimeout:    150 * time.Millisecond,
			tickDuration:      50 * time.Millisecond,
			expectGomegaWithT: false,
		},
		{
			name: "should retry terminating failed pipeline successfully",
			failedPipelines: map[string]pipeline.PipelineVersion{
				"test-uid-123_1": {
					Name:    "test-pipeline",
					Version: 1,
					UID:     "test-uid-123",
				},
			},
			needsLoadBalancer:     true,
			needsConflictResolver: true,
			setupMocks: func(mockPipelineHandler *mock.MockPipelineHandler, mockLoadBalancer *mock2.MockLoadBalancer, failedPipelines map[string]pipeline.PipelineVersion) {
				mockLoadBalancer.EXPECT().
					GetServersForKey("test-uid-123").
					Return([]string{})

				mockPipelineHandler.EXPECT().
					GetPipelineVersion("test-pipeline", uint32(1), "test-uid-123").
					Return(&pipeline.PipelineVersion{
						Name:    "test-pipeline",
						Version: 1,
						UID:     "test-uid-123",
						State: &pipeline.PipelineState{
							Status: pipeline.PipelineTerminating,
						},
					}, nil)

				mockPipelineHandler.EXPECT().
					SetPipelineState("test-pipeline", uint32(1), "test-uid-123", pipeline.PipelineTerminating, gomock.Any(), gomock.Any()).
					Return(nil)
			},
			contextTimeout:    150 * time.Millisecond,
			tickDuration:      50 * time.Millisecond,
			expectGomegaWithT: false,
		},
		{
			name: "should remove pipeline from failed list when not found",
			failedPipelines: map[string]pipeline.PipelineVersion{
				"test-uid-123_1": {
					Name:    "test-pipeline",
					Version: 1,
					UID:     "test-uid-123",
				},
			},
			needsLoadBalancer:     false,
			needsConflictResolver: false,
			setupMocks: func(mockPipelineHandler *mock.MockPipelineHandler, mockLoadBalancer *mock2.MockLoadBalancer, failedPipelines map[string]pipeline.PipelineVersion) {
				mockPipelineHandler.EXPECT().
					GetPipelineVersion("test-pipeline", uint32(1), "test-uid-123").
					Return(nil, &pipeline.PipelineNotFoundErr{})
			},
			contextTimeout: 150 * time.Millisecond,
			tickDuration:   50 * time.Millisecond,
			validateResult: func(g *WithT, server *ChainerServer) {
				g.Expect(server.failedDeletePipelines).ToNot(HaveKey("test-uid-123_1"))
			},
			expectGomegaWithT: true,
		},
		{
			name: "should remove pipeline from failed list on UID mismatch",
			failedPipelines: map[string]pipeline.PipelineVersion{
				"test-uid-123_1": {
					Name:    "test-pipeline",
					Version: 1,
					UID:     "test-uid-123",
				},
			},
			needsLoadBalancer:     false,
			needsConflictResolver: false,
			setupMocks: func(mockPipelineHandler *mock.MockPipelineHandler, mockLoadBalancer *mock2.MockLoadBalancer, failedPipelines map[string]pipeline.PipelineVersion) {
				mockPipelineHandler.EXPECT().
					GetPipelineVersion("test-pipeline", uint32(1), "test-uid-123").
					Return(nil, &pipeline.PipelineVersionUidMismatchErr{})
			},
			contextTimeout: 150 * time.Millisecond,
			tickDuration:   50 * time.Millisecond,
			validateResult: func(g *WithT, server *ChainerServer) {
				g.Expect(server.failedDeletePipelines).ToNot(HaveKey("test-uid-123_1"))
			},
			expectGomegaWithT: true,
		},
		{
			name: "should remove pipeline from failed list on version not found",
			failedPipelines: map[string]pipeline.PipelineVersion{
				"test-uid-123_1": {
					Name:    "test-pipeline",
					Version: 1,
					UID:     "test-uid-123",
				},
			},
			needsLoadBalancer:     false,
			needsConflictResolver: false,
			setupMocks: func(mockPipelineHandler *mock.MockPipelineHandler, mockLoadBalancer *mock2.MockLoadBalancer, failedPipelines map[string]pipeline.PipelineVersion) {
				mockPipelineHandler.EXPECT().
					GetPipelineVersion("test-pipeline", uint32(1), "test-uid-123").
					Return(nil, &pipeline.PipelineVersionNotFoundErr{})
			},
			contextTimeout: 150 * time.Millisecond,
			tickDuration:   50 * time.Millisecond,
			validateResult: func(g *WithT, server *ChainerServer) {
				g.Expect(server.failedDeletePipelines).ToNot(HaveKey("test-uid-123_1"))
			},
			expectGomegaWithT: true,
		},
		{
			name: "should continue processing on generic error",
			failedPipelines: map[string]pipeline.PipelineVersion{
				"test-uid-123_1": {
					Name:    "test-pipeline",
					Version: 1,
					UID:     "test-uid-123",
				},
			},
			needsLoadBalancer:     false,
			needsConflictResolver: false,
			setupMocks: func(mockPipelineHandler *mock.MockPipelineHandler, mockLoadBalancer *mock2.MockLoadBalancer, failedPipelines map[string]pipeline.PipelineVersion) {
				mockPipelineHandler.EXPECT().
					GetPipelineVersion("test-pipeline", uint32(1), "test-uid-123").
					Return(nil, errors.New("generic error")).
					MinTimes(1)
			},
			contextTimeout: 150 * time.Millisecond,
			tickDuration:   50 * time.Millisecond,
			validateResult: func(g *WithT, server *ChainerServer) {
				// Pipeline should still be in failed list
				g.Expect(server.failedDeletePipelines).To(HaveKey("test-uid-123_1"))
			},
			expectGomegaWithT: true,
		},
		{
			name: "should process multiple failed pipelines",
			failedPipelines: map[string]pipeline.PipelineVersion{
				"uid-1_1": {
					Name:    "pipeline-1",
					Version: 1,
					UID:     "uid-1",
				},
				"uid-2_1": {
					Name:    "pipeline-2",
					Version: 1,
					UID:     "uid-2",
				},
			},
			needsLoadBalancer:     true,
			needsConflictResolver: true,
			setupMocks: func(mockPipelineHandler *mock.MockPipelineHandler, mockLoadBalancer *mock2.MockLoadBalancer, failedPipelines map[string]pipeline.PipelineVersion) {
				mockPipelineHandler.EXPECT().
					GetPipelineVersion("pipeline-1", uint32(1), "uid-1").
					Return(&pipeline.PipelineVersion{
						Name:    "pipeline-1",
						Version: 1,
						UID:     "uid-1",
						State: &pipeline.PipelineState{
							Status: pipeline.PipelineTerminating,
						},
					}, nil)

				mockPipelineHandler.EXPECT().
					GetPipelineVersion("pipeline-2", uint32(1), "uid-2").
					Return(&pipeline.PipelineVersion{
						Name:    "pipeline-2",
						Version: 1,
						UID:     "uid-2",
						State: &pipeline.PipelineState{
							Status: pipeline.PipelineTerminating,
						},
					}, nil)

				mockPipelineHandler.EXPECT().
					SetPipelineState(gomock.Any(), gomock.Any(), gomock.Any(), pipeline.PipelineTerminating, gomock.Any(), gomock.Any()).
					Return(nil).
					Times(2)

				mockLoadBalancer.EXPECT().GetServersForKey("uid-1").Return([]string{})
				mockLoadBalancer.EXPECT().GetServersForKey("uid-2").Return([]string{})
			},
			contextTimeout:    150 * time.Millisecond,
			tickDuration:      50 * time.Millisecond,
			expectGomegaWithT: false,
		},
		{
			name: "should tick multiple times before context cancellation",
			failedPipelines: map[string]pipeline.PipelineVersion{
				"test-uid-123_1": {
					Name:    "test-pipeline",
					Version: 1,
					UID:     "test-uid-123",
				},
			},
			needsLoadBalancer:     false,
			needsConflictResolver: false,
			setupMocks: func(mockPipelineHandler *mock.MockPipelineHandler, mockLoadBalancer *mock2.MockLoadBalancer, failedPipelines map[string]pipeline.PipelineVersion) {
				// Expect at least 2 calls (multiple ticks)
				mockPipelineHandler.EXPECT().
					GetPipelineVersion("test-pipeline", uint32(1), "test-uid-123").
					Return(nil, errors.New("some error")).
					MinTimes(2)
			},
			contextTimeout:    250 * time.Millisecond,
			tickDuration:      50 * time.Millisecond,
			expectGomegaWithT: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var g *WithT
			if tt.expectGomegaWithT {
				g = NewGomegaWithT(t)
			}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPipelineHandler := mock.NewMockPipelineHandler(ctrl)
			var mockLoadBalancer *mock2.MockLoadBalancer
			if tt.needsLoadBalancer {
				mockLoadBalancer = mock2.NewMockLoadBalancer(ctrl)
			}

			// Setup mocks for this test case
			if tt.setupMocks != nil {
				tt.setupMocks(mockPipelineHandler, mockLoadBalancer, tt.failedPipelines)
			}

			server := &ChainerServer{
				logger:                log.New(),
				pipelineHandler:       mockPipelineHandler,
				failedDeletePipelines: tt.failedPipelines,
				streams:               make(map[string]*ChainerSubscription),
			}

			if tt.needsLoadBalancer {
				server.loadBalancer = mockLoadBalancer
			}

			if tt.needsConflictResolver {
				server.conflictResolutioner = cr.NewConflictResolution[pipeline.PipelineStatus](log.New())
			}

			var ctx context.Context
			var cancel context.CancelFunc

			if tt.contextTimeout == 0 {
				// Cancel immediately
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
			} else {
				ctx, cancel = context.WithTimeout(context.Background(), tt.contextTimeout)
				defer cancel()
			}

			done := make(chan bool)
			go func() {
				server.pollerFailedTerminatingPipelines(ctx, tt.tickDuration)
				done <- true
			}()

			// Calculate appropriate timeout based on context timeout
			testTimeout := tt.contextTimeout + 1*time.Second
			if testTimeout < 1*time.Second {
				testTimeout = 1 * time.Second
			}
			if testTimeout > 2*time.Second {
				testTimeout = 2 * time.Second
			}

			select {
			case <-done:
				// Test passed - function returned as expected
				if tt.validateResult != nil {
					tt.validateResult(g, server)
				}
			case <-time.After(testTimeout):
				t.Fatal("pollerFailedTerminatingPipelines did not return in time")
			}
		})
	}
}

func TestPollerFailedCreatingPipelines(t *testing.T) {
	tests := []struct {
		name              string
		failedPipelines   map[string]pipeline.PipelineVersion
		setupMocks        func(mockPipelineHandler *mock.MockPipelineHandler, mockLoadBalancer *mock2.MockLoadBalancer, failedPipelines map[string]pipeline.PipelineVersion)
		contextTimeout    time.Duration
		tickDuration      time.Duration
		validateResult    func(g *WithT, server *ChainerServer)
		expectGomegaWithT bool
	}{
		{
			name:            "should return when context is cancelled",
			failedPipelines: make(map[string]pipeline.PipelineVersion),
			setupMocks: func(mockPipelineHandler *mock.MockPipelineHandler, mockLoadBalancer *mock2.MockLoadBalancer, failedPipelines map[string]pipeline.PipelineVersion) {
				// No expectations - context cancelled before first tick
			},
			contextTimeout:    0, // Cancel immediately
			tickDuration:      100 * time.Millisecond,
			expectGomegaWithT: false,
		},
		{
			name:            "should skip processing when no failed pipelines exist",
			failedPipelines: make(map[string]pipeline.PipelineVersion),
			setupMocks: func(mockPipelineHandler *mock.MockPipelineHandler, mockLoadBalancer *mock2.MockLoadBalancer, failedPipelines map[string]pipeline.PipelineVersion) {
				// No expectations - empty map means no processing
			},
			contextTimeout:    150 * time.Millisecond,
			tickDuration:      50 * time.Millisecond,
			expectGomegaWithT: false,
		},
		{
			name: "should retry creating failed pipeline and remove from list on success",
			failedPipelines: map[string]pipeline.PipelineVersion{
				"test-uid-123_1": {
					Name:    "test-pipeline",
					Version: 1,
					UID:     "test-uid-123",
				},
			},
			setupMocks: func(mockPipelineHandler *mock.MockPipelineHandler, mockLoadBalancer *mock2.MockLoadBalancer, failedPipelines map[string]pipeline.PipelineVersion) {
				mockPipelineHandler.EXPECT().
					GetPipelineVersion("test-pipeline", uint32(1), "test-uid-123").
					Return(&pipeline.PipelineVersion{
						Name:    "test-pipeline",
						Version: 1,
						UID:     "test-uid-123",
						State: &pipeline.PipelineState{
							Status: pipeline.PipelineCreating,
						},
					}, nil)

				mockPipelineHandler.EXPECT().
					SetPipelineState("test-pipeline", uint32(1), "test-uid-123", pipeline.PipelineCreate, gomock.Any(), util.SourceChainerServer).
					Return(nil)
			},
			contextTimeout: 150 * time.Millisecond,
			tickDuration:   50 * time.Millisecond,
			validateResult: func(g *WithT, server *ChainerServer) {
				g.Expect(server.failedCreatePipelines).ToNot(HaveKey("test-uid-123_1"))
			},
			expectGomegaWithT: true,
		},
		{
			name: "should remove pipeline from failed list when not found",
			failedPipelines: map[string]pipeline.PipelineVersion{
				"test-uid-123_1": {
					Name:    "test-pipeline",
					Version: 1,
					UID:     "test-uid-123",
				},
			},
			setupMocks: func(mockPipelineHandler *mock.MockPipelineHandler, mockLoadBalancer *mock2.MockLoadBalancer, failedPipelines map[string]pipeline.PipelineVersion) {
				mockPipelineHandler.EXPECT().
					GetPipelineVersion("test-pipeline", uint32(1), "test-uid-123").
					Return(nil, &pipeline.PipelineNotFoundErr{})
			},
			contextTimeout: 150 * time.Millisecond,
			tickDuration:   50 * time.Millisecond,
			validateResult: func(g *WithT, server *ChainerServer) {
				g.Expect(server.failedCreatePipelines).ToNot(HaveKey("test-uid-123_1"))
			},
			expectGomegaWithT: true,
		},
		{
			name: "should remove pipeline from failed list on UID mismatch",
			failedPipelines: map[string]pipeline.PipelineVersion{
				"test-uid-123_1": {
					Name:    "test-pipeline",
					Version: 1,
					UID:     "test-uid-123",
				},
			},
			setupMocks: func(mockPipelineHandler *mock.MockPipelineHandler, mockLoadBalancer *mock2.MockLoadBalancer, failedPipelines map[string]pipeline.PipelineVersion) {
				mockPipelineHandler.EXPECT().
					GetPipelineVersion("test-pipeline", uint32(1), "test-uid-123").
					Return(nil, &pipeline.PipelineVersionUidMismatchErr{}).
					Times(1)
			},
			contextTimeout: 150 * time.Millisecond,
			tickDuration:   50 * time.Millisecond,
			validateResult: func(g *WithT, server *ChainerServer) {
				g.Expect(server.failedCreatePipelines).ToNot(HaveKey("test-uid-123_1"))
			},
			expectGomegaWithT: true,
		},
		{
			name: "should remove pipeline from failed list on version not found",
			failedPipelines: map[string]pipeline.PipelineVersion{
				"test-uid-123_1": {
					Name:    "test-pipeline",
					Version: 1,
					UID:     "test-uid-123",
				},
			},
			setupMocks: func(mockPipelineHandler *mock.MockPipelineHandler, mockLoadBalancer *mock2.MockLoadBalancer, failedPipelines map[string]pipeline.PipelineVersion) {
				mockPipelineHandler.EXPECT().
					GetPipelineVersion("test-pipeline", uint32(1), "test-uid-123").
					Return(nil, &pipeline.PipelineVersionNotFoundErr{}).
					Times(1)
			},
			contextTimeout: 150 * time.Millisecond,
			tickDuration:   50 * time.Millisecond,
			validateResult: func(g *WithT, server *ChainerServer) {
				g.Expect(server.failedCreatePipelines).ToNot(HaveKey("test-uid-123_1"))
			},
			expectGomegaWithT: true,
		},
		{
			name: "should keep pipeline in failed list on generic error from rebalancePipeline",
			failedPipelines: map[string]pipeline.PipelineVersion{
				"test-uid-123_1": {
					Name:    "test-pipeline",
					Version: 1,
					UID:     "test-uid-123",
				},
			},
			setupMocks: func(mockPipelineHandler *mock.MockPipelineHandler, mockLoadBalancer *mock2.MockLoadBalancer, failedPipelines map[string]pipeline.PipelineVersion) {
				mockPipelineHandler.EXPECT().
					GetPipelineVersion("test-pipeline", uint32(1), "test-uid-123").
					Return(&pipeline.PipelineVersion{
						Name:    "test-pipeline",
						Version: 1,
						UID:     "test-uid-123",
						State: &pipeline.PipelineState{
							Status: pipeline.PipelineCreating,
						},
					}, nil).
					MinTimes(1)

				mockPipelineHandler.EXPECT().
					SetPipelineState("test-pipeline", uint32(1), "test-uid-123", pipeline.PipelineCreate, gomock.Any(), util.SourceChainerServer).
					Return(errors.New("failed to set state")).
					MinTimes(1)
			},
			contextTimeout: 150 * time.Millisecond,
			tickDuration:   50 * time.Millisecond,
			validateResult: func(g *WithT, server *ChainerServer) {
				g.Expect(server.failedCreatePipelines).To(HaveKey("test-uid-123_1"))
			},
			expectGomegaWithT: true,
		},
		{
			name: "should keep pipeline in failed list on GetPipelineVersion generic error",
			failedPipelines: map[string]pipeline.PipelineVersion{
				"test-uid-123_1": {
					Name:    "test-pipeline",
					Version: 1,
					UID:     "test-uid-123",
				},
			},
			setupMocks: func(mockPipelineHandler *mock.MockPipelineHandler, mockLoadBalancer *mock2.MockLoadBalancer, failedPipelines map[string]pipeline.PipelineVersion) {
				mockPipelineHandler.EXPECT().
					GetPipelineVersion("test-pipeline", uint32(1), "test-uid-123").
					Return(nil, errors.New("database connection failed")).
					MinTimes(1)
			},
			contextTimeout: 150 * time.Millisecond,
			tickDuration:   50 * time.Millisecond,
			validateResult: func(g *WithT, server *ChainerServer) {
				g.Expect(server.failedCreatePipelines).To(HaveKey("test-uid-123_1"))
			},
			expectGomegaWithT: true,
		},
		{
			name: "should process multiple failed pipelines",
			failedPipelines: map[string]pipeline.PipelineVersion{
				"uid-1_1": {
					Name:    "pipeline-1",
					Version: 1,
					UID:     "uid-1",
				},
				"uid-2_1": {
					Name:    "pipeline-2",
					Version: 1,
					UID:     "uid-2",
				},
			},
			setupMocks: func(mockPipelineHandler *mock.MockPipelineHandler, mockLoadBalancer *mock2.MockLoadBalancer, failedPipelines map[string]pipeline.PipelineVersion) {
				mockPipelineHandler.EXPECT().
					GetPipelineVersion("pipeline-1", uint32(1), "uid-1").
					Return(&pipeline.PipelineVersion{
						Name:    "pipeline-1",
						Version: 1,
						UID:     "uid-1",
						State: &pipeline.PipelineState{
							Status: pipeline.PipelineCreating,
						},
					}, nil)

				mockPipelineHandler.EXPECT().
					GetPipelineVersion("pipeline-2", uint32(1), "uid-2").
					Return(&pipeline.PipelineVersion{
						Name:    "pipeline-2",
						Version: 1,
						UID:     "uid-2",
						State: &pipeline.PipelineState{
							Status: pipeline.PipelineCreating,
						},
					}, nil)

				mockPipelineHandler.EXPECT().
					SetPipelineState(gomock.Any(), gomock.Any(), gomock.Any(), pipeline.PipelineCreate, gomock.Any(), util.SourceChainerServer).
					Return(nil).
					Times(2)
			},
			contextTimeout: 150 * time.Millisecond,
			tickDuration:   50 * time.Millisecond,
			validateResult: func(g *WithT, server *ChainerServer) {
				g.Expect(server.failedCreatePipelines).ToNot(HaveKey("uid-1_1"))
				g.Expect(server.failedCreatePipelines).ToNot(HaveKey("uid-2_1"))
			},
			expectGomegaWithT: true,
		},
		{
			name: "should process mixed success and failure scenarios",
			failedPipelines: map[string]pipeline.PipelineVersion{
				"uid-success_1": {
					Name:    "pipeline-success",
					Version: 1,
					UID:     "uid-success",
				},
				"uid-fail_1": {
					Name:    "pipeline-fail",
					Version: 1,
					UID:     "uid-fail",
				},
				"uid-notfound_1": {
					Name:    "pipeline-notfound",
					Version: 1,
					UID:     "uid-notfound",
				},
			},
			setupMocks: func(mockPipelineHandler *mock.MockPipelineHandler, mockLoadBalancer *mock2.MockLoadBalancer, failedPipelines map[string]pipeline.PipelineVersion) {
				// Success case
				mockPipelineHandler.EXPECT().
					GetPipelineVersion("pipeline-success", uint32(1), "uid-success").
					Return(&pipeline.PipelineVersion{
						Name:    "pipeline-success",
						Version: 1,
						UID:     "uid-success",
						State:   &pipeline.PipelineState{Status: pipeline.PipelineCreating},
					}, nil).
					MinTimes(1)

				mockPipelineHandler.EXPECT().
					SetPipelineState("pipeline-success", uint32(1), "uid-success", pipeline.PipelineCreate, gomock.Any(), util.SourceChainerServer).
					Return(nil).
					MinTimes(1)

				// Failure case
				mockPipelineHandler.EXPECT().
					GetPipelineVersion("pipeline-fail", uint32(1), "uid-fail").
					Return(&pipeline.PipelineVersion{
						Name:    "pipeline-fail",
						Version: 1,
						UID:     "uid-fail",
						State:   &pipeline.PipelineState{Status: pipeline.PipelineCreating},
					}, nil).
					MinTimes(1)

				mockPipelineHandler.EXPECT().
					SetPipelineState("pipeline-fail", uint32(1), "uid-fail", pipeline.PipelineCreate, gomock.Any(), util.SourceChainerServer).
					Return(errors.New("state update failed")).
					MinTimes(1)

				// Not found case
				mockPipelineHandler.EXPECT().
					GetPipelineVersion("pipeline-notfound", uint32(1), "uid-notfound").
					Return(nil, &pipeline.PipelineNotFoundErr{}).
					MinTimes(1)
			},
			contextTimeout: 150 * time.Millisecond,
			tickDuration:   50 * time.Millisecond,
			validateResult: func(g *WithT, server *ChainerServer) {
				// Success and notfound should be removed
				g.Expect(server.failedCreatePipelines).ToNot(HaveKey("uid-success_1"))
				g.Expect(server.failedCreatePipelines).ToNot(HaveKey("uid-notfound_1"))
				// Failure should remain
				g.Expect(server.failedCreatePipelines).To(HaveKey("uid-fail_1"))
			},
			expectGomegaWithT: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var g *WithT
			if tt.expectGomegaWithT {
				g = NewGomegaWithT(t)
			}

			ctrl := gomock.NewController(t)

			mockPipelineHandler := mock.NewMockPipelineHandler(ctrl)
			mockLoadBalancer := mock2.NewMockLoadBalancer(ctrl)

			// Setup mocks for this test case
			if tt.setupMocks != nil {
				tt.setupMocks(mockPipelineHandler, mockLoadBalancer, tt.failedPipelines)
			}

			server := &ChainerServer{
				logger:                log.New(),
				pipelineHandler:       mockPipelineHandler,
				loadBalancer:          mockLoadBalancer,
				failedCreatePipelines: tt.failedPipelines,
				streams:               make(map[string]*ChainerSubscription),
			}

			var ctx context.Context
			var cancel context.CancelFunc

			if tt.contextTimeout == 0 {
				// Cancel immediately
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
			} else {
				ctx, cancel = context.WithTimeout(context.Background(), tt.contextTimeout)
				defer cancel()
			}

			done := make(chan bool)
			go func() {
				server.pollerFailedCreatingPipelines(ctx, tt.tickDuration)
				done <- true
			}()

			select {
			case <-done:
				// Test passed - function returned as expected
				if tt.validateResult != nil {
					tt.validateResult(g, server)
				}
			case <-time.After(tt.contextTimeout + 1*time.Second):
				t.Fatal("pollerFailedCreatingPipelines did not return in time")
			}
		})
	}
}

func TestCreateTopicSources(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name         string
		server       *ChainerServer
		pipelineName string
		inputs       []string
		sources      []*chainer.PipelineTopic
	}

	getPtrStr := func(val string) *string { return &val }
	createTopicNamer := func(namespace string, topicPrefix string) *kafka.TopicNamer {
		tn, err := kafka.NewTopicNamer(namespace, topicPrefix)
		g.Expect(err).To(BeNil())
		return tn
	}
	tests := []test{
		{
			name: "misc inputs",
			server: &ChainerServer{
				logger:     log.New(),
				topicNamer: createTopicNamer("default", "seldon"),
			},
			pipelineName: "p1",
			inputs: []string{
				"a",
				"b.inputs",
				"c.inputs.t1",
			},
			sources: []*chainer.PipelineTopic{
				{PipelineName: "p1", TopicName: "seldon.default.model.a", Tensor: nil},
				{PipelineName: "p1", TopicName: "seldon.default.model.b.inputs", Tensor: nil},
				{PipelineName: "p1", TopicName: "seldon.default.model.c.inputs", Tensor: getPtrStr("t1")},
			},
		},
		{
			name: "default inputs",
			server: &ChainerServer{
				logger:     log.New(),
				topicNamer: createTopicNamer("ns1", "seldon"),
			},
			pipelineName: "p1",
			inputs:       []string{},
			sources: []*chainer.PipelineTopic{
				{PipelineName: "p1", TopicName: "seldon.ns1.pipeline.p1.inputs", Tensor: nil},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sources := test.server.createTopicSources(test.inputs, test.pipelineName)
			g.Expect(sources).To(Equal(test.sources))
		})
	}
}

func TestCreatePipelineTopicSources(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name    string
		server  *ChainerServer
		inputs  []string
		sources []*chainer.PipelineTopic
	}

	getPtrStr := func(val string) *string { return &val }
	createTopicNamer := func(namespace string, topicPrefix string) *kafka.TopicNamer {
		tn, err := kafka.NewTopicNamer(namespace, topicPrefix)
		g.Expect(err).To(BeNil())
		return tn
	}
	tests := []test{
		{
			name: "misc inputs",
			server: &ChainerServer{
				logger:     log.New(),
				topicNamer: createTopicNamer("default", "seldon"),
			},
			inputs: []string{
				"foo.inputs",
				"foo.outputs",
				"foo.step.bar.inputs",
				"foo.step.bar.outputs",
				"foo.step.bar.inputs.tensora",
			},
			sources: []*chainer.PipelineTopic{
				{PipelineName: "foo", TopicName: "seldon.default.pipeline.foo.inputs", Tensor: nil},
				{PipelineName: "foo", TopicName: "seldon.default.pipeline.foo.outputs", Tensor: nil},
				{PipelineName: "foo", TopicName: "seldon.default.model.bar.inputs", Tensor: nil},
				{PipelineName: "foo", TopicName: "seldon.default.model.bar.outputs", Tensor: nil},
				{PipelineName: "foo", TopicName: "seldon.default.model.bar.inputs", Tensor: getPtrStr("tensora")},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sources := test.server.createPipelineTopicSources(test.inputs)
			g.Expect(sources).To(Equal(test.sources))
		})
	}
}

func TestCreateTriggerSources(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name         string
		server       *ChainerServer
		pipelineName string
		inputs       []string
		sources      []*chainer.PipelineTopic
	}

	createTopicNamer := func(namespace string, topicPrefix string) *kafka.TopicNamer {
		tn, err := kafka.NewTopicNamer(namespace, topicPrefix)
		g.Expect(err).To(BeNil())
		return tn
	}
	getPtrStr := func(val string) *string { return &val }
	tests := []test{
		{
			name: "misc inputs",
			server: &ChainerServer{
				logger:     log.New(),
				topicNamer: createTopicNamer("default", "seldon"),
			},
			pipelineName: "p1",
			inputs: []string{
				"a",
				"b.inputs",
				"c.inputs.t1",
			},
			sources: []*chainer.PipelineTopic{
				{PipelineName: "p1", TopicName: "seldon.default.model.a", Tensor: nil},
				{PipelineName: "p1", TopicName: "seldon.default.model.b.inputs", Tensor: nil},
				{PipelineName: "p1", TopicName: "seldon.default.model.c.inputs", Tensor: getPtrStr("t1")},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sources := test.server.createTriggerSources(test.inputs, test.pipelineName)
			g.Expect(sources).To(Equal(test.sources))
		})
	}
}

// test to make sure we remove old versions of the pipeline when a new version is added
func TestPipelineRollingUpgradeEvents(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name       string
		loadReqV1  *scheduler.Pipeline
		loadReqV2  *scheduler.Pipeline
		err        bool // when true old version was not marked as ready
		connection bool
		ctx        context.Context
	}

	tests := []test{
		{
			name: "old version removed - was ready",
			loadReqV1: &scheduler.Pipeline{

				Name:    "foo",
				Version: 1,
				Uid:     "x",
				Steps: []*scheduler.PipelineStep{
					{
						Name: "a",
					},
				},
			},
			loadReqV2: &scheduler.Pipeline{

				Name:    "foo",
				Version: 2,
				Uid:     "x",
				Steps: []*scheduler.PipelineStep{
					{
						Name: "a",
					},
				},
			},
			err:        false,
			connection: true,
			ctx:        context.Background(),
		},
		{
			name: "old version removed - was not ready",
			loadReqV1: &scheduler.Pipeline{

				Name:    "foo",
				Version: 1,
				Uid:     "x",
				Steps: []*scheduler.PipelineStep{
					{
						Name: "a",
					},
				},
			},
			loadReqV2: &scheduler.Pipeline{

				Name:    "foo",
				Version: 2,
				Uid:     "x",
				Steps: []*scheduler.PipelineStep{
					{
						Name: "a",
					},
				},
			},
			err:        true,
			connection: true,
			ctx:        context.Background(),
		},
		{
			name: "no new version",
			loadReqV1: &scheduler.Pipeline{

				Name:    "foo",
				Version: 1,
				Uid:     "x",
				Steps: []*scheduler.PipelineStep{
					{
						Name: "a",
					},
				},
			},
			err:        false,
			connection: true,
			ctx:        context.Background(),
		},
		{
			name: "no connection",
			loadReqV1: &scheduler.Pipeline{

				Name:    "foo",
				Version: 1,
				Uid:     "x",
				Steps: []*scheduler.PipelineStep{
					{
						Name: "a",
					},
				},
			},
			loadReqV2: &scheduler.Pipeline{

				Name:    "foo",
				Version: 2,
				Uid:     "x",
				Steps: []*scheduler.PipelineStep{
					{
						Name: "a",
					},
				},
			},
			err:        false,
			connection: false,
			ctx:        context.Background(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			serverName := "dummy"
			s, _ := createTestScheduler(t, serverName)

			stream := newStubServerStatusServer(10, test.ctx)
			if test.connection {
				s.mu.Lock()
				s.streams[serverName] = &ChainerSubscription{
					name:   "dummy",
					stream: stream,
					fin:    make(chan bool),
				}
				g.Expect(s.streams[serverName]).ToNot(BeNil())
				s.mu.Unlock()
			}

			err := s.pipelineHandler.AddPipeline(test.loadReqV1) // version 1
			g.Expect(err).To(BeNil())

			// to allow events to propagate
			time.Sleep(500 * time.Millisecond)

			// read create event
			if test.connection {
				var psr *chainer.PipelineUpdateMessage
				select {
				case next := <-stream.msgs:
					psr = next
				case <-time.After(2 * time.Second):
					t.Fail()
				}
				g.Expect(psr).ToNot(BeNil())
				g.Expect(psr.Pipeline).To(Equal(test.loadReqV1.Name))
				g.Expect(psr.Version).To(Equal(uint32(test.loadReqV1.Version)))
				g.Expect(psr.Op).To(Equal(chainer.PipelineUpdateMessage_Create))

				if !test.err {
					// simulate an ack from dataflow-engine
					err = s.pipelineHandler.SetPipelineState(test.loadReqV1.Name, test.loadReqV1.Version, test.loadReqV1.Uid, pipeline.PipelineReady, "", util.SourceChainerServer)
					g.Expect(err).To(BeNil())
				}
			}

			if test.loadReqV2 != nil {
				err = s.pipelineHandler.AddPipeline(test.loadReqV2) // version 2
				g.Expect(err).To(BeNil())

				// to allow events to propagate
				time.Sleep(500 * time.Millisecond)

				// read new create event
				if test.connection {
					var psr *chainer.PipelineUpdateMessage
					select {
					case next := <-stream.msgs:
						psr = next
					case <-time.After(2 * time.Second):
						t.Fail()
					}
					g.Expect(psr).ToNot(BeNil())
					g.Expect(psr.Pipeline).To(Equal(test.loadReqV2.Name))
					g.Expect(psr.Version).To(Equal(uint32(test.loadReqV2.Version)))
					g.Expect(psr.Op).To(Equal(chainer.PipelineUpdateMessage_Create))

					if !test.err {
						// simulate an ack from dataflow-engine
						err = s.pipelineHandler.SetPipelineState(test.loadReqV2.Name, test.loadReqV2.Version, test.loadReqV2.Uid, pipeline.PipelineReady, "", util.SourceChainerServer)
						g.Expect(err).To(BeNil())
					}
				}
			}

			if test.connection {
				if test.loadReqV2 != nil && !test.err {
					// read delete event for old version - this event is only triggered
					// after the pipeline is marked as ready
					var psr *chainer.PipelineUpdateMessage
					select {
					case next := <-stream.msgs:
						psr = next
					case <-time.After(2 * time.Second):
						t.Fail()
					}

					g.Expect(psr).ToNot(BeNil())
					g.Expect(psr.Pipeline).To(Equal(test.loadReqV1.Name))
					g.Expect(psr.Version).To(Equal(uint32(test.loadReqV1.Version)))
					g.Expect(psr.Op).To(Equal(chainer.PipelineUpdateMessage_Delete))
				}
			} else {
				// in this case we have a rolling update to a new version but the connection
				// to dataflow-engine is not available so we should have an error
				pipeline, err := s.pipelineHandler.GetPipeline(test.loadReqV2.Name)
				g.Expect(err).To(BeNil())
				// error message should be set
				g.Expect(pipeline.GetLatestPipelineVersion().State.Reason).To(Equal("no dataflow engines available to handle pipeline"))
			}

		})
	}
}

func TestPipelineEvents(t *testing.T) {

	type test struct {
		name                string
		loadReq             *scheduler.Pipeline
		status              pipeline.PipelineStatus
		connection          bool
		expectedEventStatus chainer.PipelineUpdateMessage_PipelineOperation
	}

	tests := []test{
		{
			name: "add new pipeline version",
			loadReq: &scheduler.Pipeline{

				Name:    "foo",
				Version: 1,
				Uid:     "x",
				Steps: []*scheduler.PipelineStep{
					{
						Name: "a",
					},
				},
			},
			status:              pipeline.PipelineCreate,
			expectedEventStatus: chainer.PipelineUpdateMessage_Create,
			connection:          true,
		},
		{
			name: "remove pipeline version",
			loadReq: &scheduler.Pipeline{

				Name:    "foo",
				Version: 1,
				Uid:     "x",
				Steps: []*scheduler.PipelineStep{
					{
						Name: "a",
					},
				},
			},
			status:              pipeline.PipelineTerminate,
			expectedEventStatus: chainer.PipelineUpdateMessage_Delete,
			connection:          true,
		},
		{
			name: "no connection",
			loadReq: &scheduler.Pipeline{

				Name:    "foo",
				Version: 1,
				Uid:     "x",
				Steps: []*scheduler.PipelineStep{
					{
						Name: "a",
					},
				},
			},
			status:     pipeline.PipelineTerminate,
			connection: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			serverName := "dummy"
			s, _ := createTestScheduler(t, serverName)

			stream := newStubServerStatusServer(10, context.Background())

			err := s.pipelineHandler.AddPipeline(test.loadReq) // version 1
			g.Expect(err).To(BeNil())

			// wait for the event to propagate to point ChainerServer realises there's no subscribers and discards the
			// "add pipeline" event
			time.Sleep(500 * time.Millisecond)

			if test.connection {
				s.mu.Lock()
				s.streams[serverName] = &ChainerSubscription{
					name:   "dummy",
					stream: stream,
					fin:    make(chan bool),
				}
				g.Expect(s.streams[serverName]).ToNot(BeNil())
				s.mu.Unlock()
			}

			err = s.pipelineHandler.SetPipelineState(test.loadReq.Name, test.loadReq.Version, test.loadReq.Uid, test.status, "", "some-source")
			g.Expect(err).To(BeNil())

			if test.connection {
				var psr *chainer.PipelineUpdateMessage
				select {
				case next := <-stream.msgs:
					psr = next
				case <-time.After(2 * time.Second):
					t.Fail()
				}

				g.Expect(psr).ToNot(BeNil())
				g.Expect(psr.Pipeline).To(Equal(test.loadReq.Name))
				g.Expect(psr.Version).To(Equal(test.loadReq.Version))
				g.Expect(psr.Op).To(Equal(test.expectedEventStatus))
				return
			}

			// wait for events to be processed
			time.Sleep(time.Second)
			// in this case we do not have a dataflow-engine connection so we should have an error message
			pipeline, err := s.pipelineHandler.GetPipeline(test.loadReq.Name)
			g.Expect(err).To(BeNil())
			// error message should be set
			g.Expect(pipeline.GetLatestPipelineVersion().State.Reason).To(Equal("no dataflow engines available to handle pipeline"))
		})
	}
}

func TestPipelineRebalance(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name           string
		loadReq        *scheduler.Pipeline
		status         pipeline.PipelineStatus
		expectedStatus pipeline.PipelineStatus
		expectedOp     chainer.PipelineUpdateMessage_PipelineOperation
		connection     bool
	}

	tests := []test{
		{
			name: "rebalance ready pipeline",
			loadReq: &scheduler.Pipeline{

				Name:    "foo",
				Version: 1,
				Uid:     "x",
				Steps: []*scheduler.PipelineStep{
					{
						Name: "a",
					},
				},
			},
			status:         pipeline.PipelineReady,
			expectedStatus: pipeline.PipelineCreating,
			expectedOp:     chainer.PipelineUpdateMessage_Create,
			connection:     true,
		},
		{
			name: "rebalance creating pipeline",
			loadReq: &scheduler.Pipeline{

				Name:    "foo",
				Version: 1,
				Uid:     "x",
				Steps: []*scheduler.PipelineStep{
					{
						Name: "a",
					},
				},
			},
			status:         pipeline.PipelineCreating,
			expectedStatus: pipeline.PipelineCreating,
			expectedOp:     chainer.PipelineUpdateMessage_Create,
			connection:     true,
		},
		{
			name: "rebalance terminating pipeline",
			loadReq: &scheduler.Pipeline{

				Name:    "foo",
				Version: 1,
				Uid:     "x",
				Steps: []*scheduler.PipelineStep{
					{
						Name: "a",
					},
				},
			},
			status:         pipeline.PipelineTerminating,
			expectedStatus: pipeline.PipelineTerminating,
			expectedOp:     chainer.PipelineUpdateMessage_Delete,
			connection:     true,
		},
		{
			name: "rebalance terminating pipeline - no connection",
			loadReq: &scheduler.Pipeline{

				Name:    "foo",
				Version: 1,
				Uid:     "x",
				Steps: []*scheduler.PipelineStep{
					{
						Name: "a",
					},
				},
			},
			status:         pipeline.PipelineTerminating,
			expectedStatus: pipeline.PipelineTerminated,
			connection:     false,
		},
		{
			name: "no connection",
			loadReq: &scheduler.Pipeline{

				Name:    "foo",
				Version: 1,
				Uid:     "x",
				Steps: []*scheduler.PipelineStep{
					{
						Name: "a",
					},
				},
			},
			status:         pipeline.PipelineReady,
			expectedStatus: pipeline.PipelineCreate,
			connection:     false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			serverName := "dummy"
			s, _ := createTestScheduler(t, serverName)
			stream := newStubServerStatusServer(20, context.Background())
			if test.connection {
				s.streams[serverName] = &ChainerSubscription{
					name:   "dummy",
					stream: stream,
					fin:    make(chan bool),
				}
				g.Expect(s.streams[serverName]).ToNot(BeNil())
			}

			err := s.pipelineHandler.AddPipeline(test.loadReq) // version 1
			g.Expect(err).To(BeNil())

			//to allow events to propagate
			time.Sleep(500 * time.Millisecond)

			err = s.pipelineHandler.SetPipelineState(test.loadReq.Name, test.loadReq.Version, test.loadReq.Uid, test.status, "", util.SourceChainerServer)
			g.Expect(err).To(BeNil())

			//to allow events to propagate
			time.Sleep(500 * time.Millisecond)

			s.rebalance()
			actualPipelineVersion, _ := s.pipelineHandler.GetPipelineVersion(test.loadReq.Name, test.loadReq.Version, test.loadReq.Uid)

			// to allow events to propagate
			time.Sleep(500 * time.Millisecond)

			if test.connection {
				var psr *chainer.PipelineUpdateMessage
				select {
				case <-stream.msgs: // skip the first message for AddPipeline
					next := <-stream.msgs
					psr = next
				default:
					t.Fail()
				}

				g.Expect(psr).ToNot(BeNil())
				g.Expect(psr.Pipeline).To(Equal(test.loadReq.Name))
				g.Expect(psr.Version).To(Equal(uint32(test.loadReq.Version)))
				g.Expect(psr.Op).To(Equal(test.expectedOp))
			} else {
				// error message should be set
				g.Expect(actualPipelineVersion.State.Reason).To(Equal("no dataflow engines available to handle pipeline"))
			}
			g.Expect(actualPipelineVersion.State.Status).To(Equal(test.expectedStatus))

		})
	}
}

func TestPipelineSubscribe(t *testing.T) {
	g := NewGomegaWithT(t)
	type ag struct {
		id      uint32
		doClose bool
	}

	type test struct {
		name                          string
		agents                        []ag
		expectedAgentsCount           int
		expectedAgentsCountAfterClose int
	}

	tests := []test{
		{
			name: "single connection",
			agents: []ag{
				{id: 1, doClose: true},
			},
			expectedAgentsCount:           1,
			expectedAgentsCountAfterClose: 0,
		},
		{
			name: "multiple connection - one not closed",
			agents: []ag{
				{id: 1, doClose: false}, {id: 2, doClose: true},
			},
			expectedAgentsCount:           2,
			expectedAgentsCountAfterClose: 1,
		},
		{
			name: "multiple connection - not closed",
			agents: []ag{
				{id: 1, doClose: false}, {id: 2, doClose: false},
			},
			expectedAgentsCount:           2,
			expectedAgentsCountAfterClose: 2,
		},
		{
			name: "multiple connection - closed",
			agents: []ag{
				{id: 1, doClose: true}, {id: 2, doClose: true},
			},
			expectedAgentsCount:           2,
			expectedAgentsCountAfterClose: 0,
		},
		{
			name: "multiple connection - duplicate",
			agents: []ag{
				{id: 1, doClose: true}, {id: 1, doClose: true}, {id: 1, doClose: true},
			},
			expectedAgentsCount:           1,
			expectedAgentsCountAfterClose: 0,
		},
		{
			name: "multiple connection - duplicate not closed",
			agents: []ag{
				{id: 1, doClose: true}, {id: 1, doClose: false}, {id: 1, doClose: true},
			},
			expectedAgentsCount:           1,
			expectedAgentsCountAfterClose: 1,
		},
	}

	getStream := func(id uint32, context context.Context, port int) *grpc.ClientConn {
		conn, _ := grpc.NewClient(fmt.Sprintf(":%d", port), grpc.WithTransportCredentials(insecure.NewCredentials()))
		grpcClient := chainer.NewChainerClient(conn)
		_, _ = grpcClient.SubscribePipelineUpdates(
			context,
			&chainer.PipelineSubscriptionRequest{
				Name: fmt.Sprintf("agent-%d", id),
			},
		)
		return conn
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			serverName := "dummy"
			s, _ := createTestScheduler(t, serverName)
			port, err := testing_utils.GetFreePortForTest()
			if err != nil {
				t.Fatal(err)
			}
			go func() {
				_ = s.StartGrpcServer(context.Background(), time.Minute, time.Minute, uint(port))
			}()

			time.Sleep(100 * time.Millisecond)

			mu := sync.Mutex{}
			streams := make([]*grpc.ClientConn, 0)
			for _, a := range test.agents {
				go func(id uint32) {
					conn := getStream(id, context.Background(), port)
					mu.Lock()
					streams = append(streams, conn)
					mu.Unlock()
				}(a.id)
			}

			maxCount := 10
			count := 0
			for count < maxCount {
				s.mu.Lock()
				if len(s.streams) == test.expectedAgentsCount {
					s.mu.Unlock()
					break
				}
				s.mu.Unlock()

				time.Sleep(100 * time.Millisecond)
				count++
			}

			s.mu.Lock()
			g.Expect(len(s.streams)).To(Equal(test.expectedAgentsCount))
			s.mu.Unlock()

			mu.Lock()
			for idx, s := range streams {
				go func(idx int, s *grpc.ClientConn) {
					if test.agents[idx].doClose {
						s.Close()
					}
				}(idx, s)
			}
			mu.Unlock()

			for count < maxCount {
				s.mu.Lock()
				if len(s.streams) == test.expectedAgentsCountAfterClose {
					s.mu.Unlock()
					break
				}

				s.mu.Unlock()
				time.Sleep(100 * time.Millisecond)
				count++
			}

			s.mu.Lock()
			g.Expect(len(s.streams)).To(Equal(test.expectedAgentsCountAfterClose))
			s.mu.Unlock()

			s.StopSendPipelineEvents()
		})
	}
}

type stubChainerServer struct {
	msgs chan *chainer.PipelineUpdateMessage
	ctx  context.Context
	grpc.ServerStream
}

var _ chainer.Chainer_SubscribePipelineUpdatesServer = (*stubChainerServer)(nil)

func newStubServerStatusServer(capacity int, ctx context.Context) *stubChainerServer {
	return &stubChainerServer{
		ctx:  ctx,
		msgs: make(chan *chainer.PipelineUpdateMessage, capacity),
	}
}

func (s *stubChainerServer) Context() context.Context {
	return s.ctx
}

func (s *stubChainerServer) Send(r *chainer.PipelineUpdateMessage) error {
	s.msgs <- r
	return nil
}

// TODO: this function is defined elsewhere, refactor to avoid duplication
func createTestScheduler(t *testing.T, serverName string) (*ChainerServer, *coordinator.EventHub) {
	logger := log.New()
	logger.SetLevel(log.DebugLevel)

	eventHub, _ := coordinator.NewEventHub(logger)

	schedulerStore := store.NewMemoryStore(logger, store.NewLocalSchedulerStore(), eventHub)
	pipelineServer := pipeline.NewPipelineStore(logger, eventHub, schedulerStore)

	data :=
		`
	{
	  "bootstrap.servers":"kafka:9092",
	  "consumer":{"session.timeout.ms": 6000, "someBool": true, "someString":"foo"},
	  "producer": {"linger.ms":0},
	  "streams": {"replication.factor": 1}
	}
	`
	configFilePath := fmt.Sprintf("%s/kafka.json", t.TempDir())
	_ = os.WriteFile(configFilePath, []byte(data), 0644)
	kc, _ := kafka_config.NewKafkaConfig(configFilePath, "debug")
	b := util.NewRingLoadBalancer(1)
	b.AddServer(serverName)
	s, _ := NewChainerServer(logger, eventHub, pipelineServer, "test-ns", b, kc, nil)

	return s, eventHub
}
