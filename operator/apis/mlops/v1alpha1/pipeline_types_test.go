package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/gomega"
	scheduler "github.com/seldonio/seldon-core/operatorv2/scheduler/apis/mlops/scheduler"
)

func TestAsPipelineDetails(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name     string
		pipeline *Pipeline
		proto    *scheduler.Pipeline
	}

	getUintPtr := func(val uint32) *uint32 { return &val }
	getJoinPtr := func(val JoinType) *JoinType { return &val }
	tests := []test{
		{
			name: "basic",
			pipeline: &Pipeline{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "foo",
					Namespace:  "default",
					Generation: 1,
				},
				Spec: PipelineSpec{
					Steps: []PipelineStep{
						{
							Name: "a",
						},
						{
							Name:   "b",
							Inputs: []string{"a"},
						},
						{
							Name:           "c",
							Inputs:         []string{"b"},
							JoinWindowMs:   getUintPtr(20),
							InputsJoinType: getJoinPtr(JoinTypeInner),
							Batch: &PipelineBatch{
								Size:     getUintPtr(100),
								WindowMs: getUintPtr(1000),
								Rolling:  true,
							},
						},
					},
					Output: &PipelineOutput{
						Steps:        []string{"c"},
						JoinWindowMs: 2,
						StepsJoin:    getJoinPtr(JoinTypeAny),
					},
				},
			},
			proto: &scheduler.Pipeline{
				Name: "foo",
				Steps: []*scheduler.PipelineStep{
					{
						Name: "a",
					},
					{
						Name:   "b",
						Inputs: []string{"a"},
					},
					{
						Name:         "c",
						Inputs:       []string{"b"},
						JoinWindowMs: getUintPtr(20),
						InputsJoin:   scheduler.PipelineStep_INNER,
						Batch: &scheduler.Batch{
							Size:     getUintPtr(100),
							WindowMs: getUintPtr(1000),
						},
					},
				},
				Output: &scheduler.PipelineOutput{
					Steps:        []string{"c"},
					JoinWindowMs: 2,
					StepsJoin:    scheduler.PipelineOutput_ANY,
				},
				KubernetesMeta: &scheduler.KubernetesMeta{
					Namespace:  "default",
					Generation: 1,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			proto := test.pipeline.AsSchedulerPipeline()
			g.Expect(proto).To(Equal(test.proto))
		})
	}
}
