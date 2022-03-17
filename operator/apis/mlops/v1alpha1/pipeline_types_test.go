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
							Name:               "c",
							Inputs:             []string{"b"},
							JoinWindowMs:       getUintPtr(20),
							PassEmptyResponses: true,
						},
					},
					Output: &PipelineOutput{
						Inputs:       []string{"c"},
						JoinWindowMs: 2,
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
						Name:               "c",
						Inputs:             []string{"b"},
						JoinWindowMs:       getUintPtr(20),
						PassEmptyResponses: true,
					},
				},
				Output: &scheduler.PipelineOutput{
					Inputs:       []string{"c"},
					JoinWindowMs: 2,
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
