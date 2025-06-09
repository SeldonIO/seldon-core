/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package dataflow

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus/hooks/test"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/chainer"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"
)

func newTestConflictResolutioner() *ConflictResolutioner {
	logger, _ := test.NewNullLogger()
	return NewConflictResolution(logger)
}

func TestCreateNewIteration(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	tests := []struct {
		name          string
		pipelineName  string
		servers       []string
		expectedClock uint64
	}{
		{
			name:          "single pipeline creation",
			pipelineName:  "pipeline1",
			servers:       []string{"a", "b", "c"},
			expectedClock: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cr := newTestConflictResolutioner()
			cr.CreateNewIteration(test.pipelineName, test.servers)

			g.Expect(cr.vectorClock[test.pipelineName]).To(gomega.Equal(test.expectedClock))
			for _, server := range test.servers {
				g.Expect(cr.vectorResponseStatus[test.pipelineName][server]).To(gomega.Equal(pipeline.PipelineStatusUnknown))
			}
		})
	}
}

func TestUpdatePipelineStatus(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	t.Run("should update status for a stream", func(t *testing.T) {
		cr := newTestConflictResolutioner()
		cr.CreateNewIteration("pipeline1", []string{"a"})

		cr.UpdatePipelineStatus("pipeline1", "a", pipeline.PipelineReady)
		g.Expect(cr.vectorResponseStatus["pipeline1"]["a"]).To(gomega.Equal(pipeline.PipelineReady))
	})
}

func TestDeletePipeline(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	t.Run("should delete vector clock and statuses", func(t *testing.T) {
		cr := newTestConflictResolutioner()
		cr.CreateNewIteration("pipeline1", []string{"a"})
		cr.DeletePipeline("pipeline1")

		_, exists1 := cr.vectorClock["pipeline1"]
		_, exists2 := cr.vectorResponseStatus["pipeline1"]

		g.Expect(exists1).To(gomega.BeFalse())
		g.Expect(exists2).To(gomega.BeFalse())
	})
}

func TestIsMessageOutdated(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	tests := []struct {
		name       string
		prepare    func(cr *ConflictResolutioner)
		message    *chainer.PipelineUpdateStatusMessage
		expectTrue bool
	}{
		{
			name: "timestamp mismatch",
			prepare: func(cr *ConflictResolutioner) {
				cr.CreateNewIteration("p1", []string{"a"})
			},
			message: &chainer.PipelineUpdateStatusMessage{
				Update: &chainer.PipelineUpdateMessage{
					Timestamp: 999,
					Pipeline:  "p1",
					Stream:    "a",
				},
			},
			expectTrue: true,
		},
		{
			name: "unknown stream",
			prepare: func(cr *ConflictResolutioner) {
				cr.CreateNewIteration("p1", []string{"a"})
			},
			message: &chainer.PipelineUpdateStatusMessage{
				Update: &chainer.PipelineUpdateMessage{
					Timestamp: 1,
					Pipeline:  "p1",
					Stream:    "b",
				},
			},
			expectTrue: true,
		},
		{
			name: "message is not outdated",
			prepare: func(cr *ConflictResolutioner) {
				cr.CreateNewIteration("p1", []string{"a"})
			},
			message: &chainer.PipelineUpdateStatusMessage{
				Update: &chainer.PipelineUpdateMessage{
					Timestamp: 1,
					Pipeline:  "p1",
					Stream:    "a",
				},
			},
			expectTrue: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cr := newTestConflictResolutioner()
			test.prepare(cr)

			result := cr.IsMessageOutdated(test.message)
			g.Expect(result).To(gomega.Equal(test.expectTrue))
		})
	}
}

func TestGetPipelineStatus(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	tests := []struct {
		name     string
		op       chainer.PipelineUpdateMessage_PipelineOperation
		statuses map[string]pipeline.PipelineStatus
		expected pipeline.PipelineStatus
	}{
		{
			name: "create creating",
			op:   chainer.PipelineUpdateMessage_Create,
			statuses: map[string]pipeline.PipelineStatus{
				"a": pipeline.PipelineReady,
				"b": pipeline.PipelineStatusUnknown,
			},
			expected: pipeline.PipelineCreating,
		},
		{
			name: "create ready (all ready)",
			op:   chainer.PipelineUpdateMessage_Create,
			statuses: map[string]pipeline.PipelineStatus{
				"a": pipeline.PipelineReady,
				"b": pipeline.PipelineReady,
			},
			expected: pipeline.PipelineReady,
		},
		{
			name: "create creating (some ready)",
			op:   chainer.PipelineUpdateMessage_Create,
			statuses: map[string]pipeline.PipelineStatus{
				"a": pipeline.PipelineReady,
				"b": pipeline.PipelineFailed,
			},
			expected: pipeline.PipelineReady,
		},
		{
			name: "create failed",
			op:   chainer.PipelineUpdateMessage_Create,
			statuses: map[string]pipeline.PipelineStatus{
				"a": pipeline.PipelineFailed,
				"b": pipeline.PipelineFailed,
			},
			expected: pipeline.PipelineFailed,
		},
		{
			name: "delete terminating",
			op:   chainer.PipelineUpdateMessage_Delete,
			statuses: map[string]pipeline.PipelineStatus{
				"a": pipeline.PipelineTerminated,
				"b": pipeline.PipelineStatusUnknown,
			},
			expected: pipeline.PipelineTerminating,
		},
		{
			name: "delete failed",
			op:   chainer.PipelineUpdateMessage_Delete,
			statuses: map[string]pipeline.PipelineStatus{
				"a": pipeline.PipelineFailed,
			},
			expected: pipeline.PipelineFailed,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cr := newTestConflictResolutioner()
			streams := []string{}
			for stream := range test.statuses {
				streams = append(streams, stream)
			}
			cr.CreateNewIteration("p1", streams)
			for s, st := range test.statuses {
				cr.UpdatePipelineStatus("p1", s, st)
			}

			msg := &chainer.PipelineUpdateStatusMessage{
				Update: &chainer.PipelineUpdateMessage{
					Op:       test.op,
					Pipeline: "p1",
				},
			}

			status, _ := cr.GetPipelineStatus("p1", msg)
			g.Expect(status).To(gomega.Equal(test.expected))
		})
	}
}
