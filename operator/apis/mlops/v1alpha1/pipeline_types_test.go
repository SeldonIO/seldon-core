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

package v1alpha1

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	scheduler "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
)

func TestAsPipelineDetails(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name     string
		pipeline *Pipeline
		proto    *scheduler.Pipeline
	}

	getUintPtr := func(val uint32) *uint32 { return &val }
	getIntPtr := func(val int32) *int32 { return &val }
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
							Name:          "b",
							Inputs:        []string{"a"},
							FilterPercent: getIntPtr(20),
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
						Name:          "b",
						Inputs:        []string{"a"},
						FilterPercent: 20,
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
