/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package conflict_resolution

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus/hooks/test"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/chainer"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"
)

func newTestConflictResolutioner() *ConflictResolutioner[pipeline.PipelineStatus] {
	logger, _ := test.NewNullLogger()
	return NewConflictResolution[pipeline.PipelineStatus](logger)
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
			CreateNewPipelineIteration(cr, test.pipelineName, test.servers)

			g.Expect(cr.VectorClock[test.pipelineName]).To(gomega.Equal(test.expectedClock))
			for _, server := range test.servers {
				g.Expect(cr.VectorResponseStatus[test.pipelineName][server]).To(gomega.Equal(pipeline.PipelineStatusUnknown))
			}
		})
	}
}

func TestUpdatePipelineStatus(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	t.Run("should update status for a stream", func(t *testing.T) {
		cr := newTestConflictResolutioner()
		CreateNewPipelineIteration(cr, "pipeline1", []string{"a"})

		cr.UpdateStatus("pipeline1", "a", pipeline.PipelineReady)
		g.Expect(cr.VectorResponseStatus["pipeline1"]["a"]).To(gomega.Equal(pipeline.PipelineReady))
	})
}

func TestDeletePipeline(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	t.Run("should delete vector clock and statuses", func(t *testing.T) {
		cr := newTestConflictResolutioner()
		CreateNewPipelineIteration(cr, "pipeline1", []string{"a"})
		cr.Delete("pipeline1")

		_, exists1 := cr.VectorClock["pipeline1"]
		_, exists2 := cr.VectorResponseStatus["pipeline1"]

		g.Expect(exists1).To(gomega.BeFalse())
		g.Expect(exists2).To(gomega.BeFalse())
	})
}

func TestIsMessageOutdated(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	tests := []struct {
		name       string
		prepare    func(cr *ConflictResolutioner[pipeline.PipelineStatus])
		message    *chainer.PipelineUpdateStatusMessage
		expectTrue bool
	}{
		{
			name: "timestamp mismatch",
			prepare: func(cr *ConflictResolutioner[pipeline.PipelineStatus]) {
				CreateNewPipelineIteration(cr, "p1", []string{"a"})
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
			prepare: func(cr *ConflictResolutioner[pipeline.PipelineStatus]) {
				CreateNewPipelineIteration(cr, "p1", []string{"a"})
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
			prepare: func(cr *ConflictResolutioner[pipeline.PipelineStatus]) {
				CreateNewPipelineIteration(cr, "p1", []string{"a"})
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

			result := IsPipelineMessageOutdated(cr, test.message)
			g.Expect(result).To(gomega.Equal(test.expectTrue))
		})
	}
}

func TestGetPipelineStatus(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	type expect struct {
		status pipeline.PipelineStatus
		msg    string
	}

	tests := []struct {
		name     string
		op       chainer.PipelineUpdateMessage_PipelineOperation
		statuses map[string]pipeline.PipelineStatus
		expect   expect
		msg      string
	}{
		{
			name: "create creating",
			op:   chainer.PipelineUpdateMessage_Create,
			statuses: map[string]pipeline.PipelineStatus{
				"a": pipeline.PipelineReady,
				"b": pipeline.PipelineStatusUnknown,
			},
			expect: expect{status: pipeline.PipelineCreating, msg: "1/2 ready "},
		},
		{
			name: "create ready (all ready)",
			op:   chainer.PipelineUpdateMessage_Create,
			statuses: map[string]pipeline.PipelineStatus{
				"a": pipeline.PipelineReady,
				"b": pipeline.PipelineReady,
			},
			expect: expect{status: pipeline.PipelineReady, msg: "2/2 ready "},
		},
		{
			name: "create creating (some ready)",
			op:   chainer.PipelineUpdateMessage_Create,
			statuses: map[string]pipeline.PipelineStatus{
				"a": pipeline.PipelineReady,
				"b": pipeline.PipelineFailed,
			},
			expect: expect{status: pipeline.PipelineReady, msg: "1/2 ready 1/2 failed "},
		},
		{
			name: "create failed",
			op:   chainer.PipelineUpdateMessage_Create,
			statuses: map[string]pipeline.PipelineStatus{
				"a": pipeline.PipelineFailed,
				"b": pipeline.PipelineFailed,
			},
			expect: expect{status: pipeline.PipelineFailed, msg: "2/2 failed "},
		},
		{
			name: "delete terminating",
			op:   chainer.PipelineUpdateMessage_Delete,
			statuses: map[string]pipeline.PipelineStatus{
				"a": pipeline.PipelineTerminated,
				"b": pipeline.PipelineStatusUnknown,
			},
			expect: expect{status: pipeline.PipelineTerminating, msg: "1/2 terminated "},
		},
		{
			name: "delete failed",
			op:   chainer.PipelineUpdateMessage_Delete,
			statuses: map[string]pipeline.PipelineStatus{
				"a": pipeline.PipelineFailedTerminating,
			},
			expect: expect{status: pipeline.PipelineFailedTerminating, msg: "1/1 failed terminating"},
		},
		{
			name: "rebalance failed",
			op:   chainer.PipelineUpdateMessage_Ready,
			statuses: map[string]pipeline.PipelineStatus{
				"a": pipeline.PipelineFailed,
				"b": pipeline.PipelineFailed,
			},
			expect: expect{status: pipeline.PipelineFailed, msg: "2/2 failed "},
		},
		{
			name: "rebalanced",
			op:   chainer.PipelineUpdateMessage_Ready,
			statuses: map[string]pipeline.PipelineStatus{
				"a": pipeline.PipelineReady,
				"b": pipeline.PipelineReady,
			},
			expect: expect{status: pipeline.PipelineReady, msg: "2/2 ready "},
		},
		{
			name: "rebalanced (some ready)",
			op:   chainer.PipelineUpdateMessage_Ready,
			statuses: map[string]pipeline.PipelineStatus{
				"a": pipeline.PipelineReady,
				"b": pipeline.PipelineFailed,
			},
			expect: expect{status: pipeline.PipelineReady, msg: "1/2 ready 1/2 failed "},
		},
		{
			name: "rebalancing all",
			op:   chainer.PipelineUpdateMessage_Rebalance,
			statuses: map[string]pipeline.PipelineStatus{
				"a": pipeline.PipelineRebalancing,
				"b": pipeline.PipelineRebalancing,
			},
			expect: expect{status: pipeline.PipelineRebalancing, msg: "2/2 rebalancing "},
		},
		{
			name: "rebalancing some",
			op:   chainer.PipelineUpdateMessage_Rebalance,
			statuses: map[string]pipeline.PipelineStatus{
				"a": pipeline.PipelineReady,
				"b": pipeline.PipelineRebalancing,
			},
			expect: expect{status: pipeline.PipelineRebalancing, msg: "1/2 ready 1/2 rebalancing "},
		},
		{
			name: "delete failed",
			op:   chainer.PipelineUpdateMessage_Delete,
			statuses: map[string]pipeline.PipelineStatus{
				"a": pipeline.PipelineFailedTerminating,
			},
			expect: expect{status: pipeline.PipelineFailedTerminating, msg: "1/1 failed terminating"},
		},
		{
			name: "delete failed and pipeline failed to create",
			op:   chainer.PipelineUpdateMessage_Delete,
			statuses: map[string]pipeline.PipelineStatus{
				"a": pipeline.PipelineFailedTerminating,
				"b": pipeline.PipelineFailed,
			},
			expect: expect{status: pipeline.PipelineFailedTerminating, msg: "1/2 failed 1/2 failed terminating"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cr := newTestConflictResolutioner()
			streams := []string{}
			for stream := range test.statuses {
				streams = append(streams, stream)
			}
			CreateNewPipelineIteration(cr, "p1", streams)
			for s, st := range test.statuses {
				cr.UpdateStatus("p1", s, st)
			}

			msg := &chainer.PipelineUpdateStatusMessage{
				Update: &chainer.PipelineUpdateMessage{
					Op:       test.op,
					Pipeline: "p1",
				},
			}

			status, outputMsg := GetPipelineStatus(cr, "p1", msg)
			g.Expect(status).To(gomega.Equal(test.expect.status))
			g.Expect(outputMsg).To(gomega.Equal(test.expect.msg))
		})
	}
}
