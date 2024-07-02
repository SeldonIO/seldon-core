/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package v1alpha1

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
)

func TestAsSchedulerExperimentRequest(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name       string
		experiment *Experiment
		proto      *scheduler.Experiment
	}

	getStrPtr := func(val string) *string { return &val }
	tests := []test{
		{
			name: "model",
			experiment: &Experiment{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "foo",
					Namespace:  "default",
					Generation: 1,
				},
				Spec: ExperimentSpec{
					Default: getStrPtr("model1"),
					Candidates: []ExperimentCandidate{
						{
							Name:   "model1",
							Weight: 20,
						},
						{
							Name:   "model2",
							Weight: 30,
						},
					},
					Mirror: &ExperimentMirror{
						Name:    "model4",
						Percent: 40,
					},
				},
			},
			proto: &scheduler.Experiment{
				Name:         "foo",
				Default:      getStrPtr("model1"),
				ResourceType: scheduler.ResourceType_MODEL,
				Candidates: []*scheduler.ExperimentCandidate{
					{
						Name:   "model1",
						Weight: 20,
					},
					{
						Name:   "model2",
						Weight: 30,
					},
				},
				Mirror: &scheduler.ExperimentMirror{
					Name:    "model4",
					Percent: 40,
				},
				KubernetesMeta: &scheduler.KubernetesMeta{
					Namespace:  "default",
					Generation: 1,
				},
			},
		},
		{
			name: "pipeline",
			experiment: &Experiment{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "foo",
					Namespace:  "default",
					Generation: 1,
				},
				Spec: ExperimentSpec{
					Default:      getStrPtr("pipeline1"),
					ResourceType: PipelineResourceType,
					Candidates: []ExperimentCandidate{
						{
							Name:   "pipeline1",
							Weight: 20,
						},
						{
							Name:   "pipeline2",
							Weight: 30,
						},
					},
					Mirror: &ExperimentMirror{
						Name:    "pipeline4",
						Percent: 40,
					},
				},
			},
			proto: &scheduler.Experiment{
				Name:         "foo",
				Default:      getStrPtr("pipeline1"),
				ResourceType: scheduler.ResourceType_PIPELINE,
				Candidates: []*scheduler.ExperimentCandidate{
					{
						Name:   "pipeline1",
						Weight: 20,
					},
					{
						Name:   "pipeline2",
						Weight: 30,
					},
				},
				Mirror: &scheduler.ExperimentMirror{
					Name:    "pipeline4",
					Percent: 40,
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
			proto := test.experiment.AsSchedulerExperimentRequest()
			g.Expect(proto).To(Equal(test.proto))
		})
	}
}
