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

package experiment

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
)

func TestLoadModel(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name       string
		proto      *scheduler.Experiment
		experiment *Experiment
	}

	getStrPtr := func(val string) *string { return &val }
	tests := []test{
		{
			name: "model",
			proto: &scheduler.Experiment{
				Name:    "foo",
				Default: getStrPtr("model1"),
				Candidates: []*scheduler.ExperimentCandidate{
					{
						Name:   "model1",
						Weight: 20,
					},
					{
						Name:   "model3",
						Weight: 20,
					},
				},
				Mirror: &scheduler.ExperimentMirror{
					Name:    "model4",
					Percent: 80,
				},
				Config: &scheduler.ExperimentConfig{
					StickySessions: true,
				},
				KubernetesMeta: &scheduler.KubernetesMeta{
					Namespace:  "default",
					Generation: 1,
				},
			},
			experiment: &Experiment{
				Name:         "foo",
				Active:       false,
				Default:      getStrPtr("model1"),
				ResourceType: ModelResourceType,
				Candidates: []*Candidate{
					{
						Name:   "model1",
						Weight: 20,
					},
					{
						Name:   "model3",
						Weight: 20,
					},
				},
				Mirror: &Mirror{
					Name:    "model4",
					Percent: 80,
				},
				Config: &Config{
					StickySessions: true,
				},
				KubernetesMeta: &KubernetesMeta{
					Namespace:  "default",
					Generation: 1,
				},
			},
		},
		{
			name: "model candidates",
			proto: &scheduler.Experiment{
				Name: "foo",
				Candidates: []*scheduler.ExperimentCandidate{
					{
						Name:   "model1",
						Weight: 20,
					},
					{
						Name:   "model3",
						Weight: 20,
					},
				},
				Mirror: &scheduler.ExperimentMirror{
					Name:    "model4",
					Percent: 80,
				},
				Config: &scheduler.ExperimentConfig{
					StickySessions: true,
				},
				KubernetesMeta: &scheduler.KubernetesMeta{
					Namespace:  "default",
					Generation: 1,
				},
			},
			experiment: &Experiment{
				Name:         "foo",
				Active:       false,
				ResourceType: ModelResourceType,
				Candidates: []*Candidate{
					{
						Name:   "model1",
						Weight: 20,
					},
					{
						Name:   "model3",
						Weight: 20,
					},
				},
				Mirror: &Mirror{
					Name:    "model4",
					Percent: 80,
				},
				Config: &Config{
					StickySessions: true,
				},
				KubernetesMeta: &KubernetesMeta{
					Namespace:  "default",
					Generation: 1,
				},
			},
		},
		{
			name: "pipeline",
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
						Weight: 20,
					},
				},
				Mirror: &scheduler.ExperimentMirror{
					Name:    "pipeline4",
					Percent: 80,
				},
				Config: &scheduler.ExperimentConfig{
					StickySessions: true,
				},
				KubernetesMeta: &scheduler.KubernetesMeta{
					Namespace:  "default",
					Generation: 1,
				},
			},
			experiment: &Experiment{
				Name:         "foo",
				Active:       false,
				Default:      getStrPtr("pipeline1"),
				ResourceType: PipelineResourceType,
				Candidates: []*Candidate{
					{
						Name:   "pipeline1",
						Weight: 20,
					},
					{
						Name:   "pipeline2",
						Weight: 20,
					},
				},
				Mirror: &Mirror{
					Name:    "pipeline4",
					Percent: 80,
				},
				Config: &Config{
					StickySessions: true,
				},
				KubernetesMeta: &KubernetesMeta{
					Namespace:  "default",
					Generation: 1,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			experiment := CreateExperimentFromRequest(test.proto)
			g.Expect(experiment).To(Equal(test.experiment))
		})
	}

}
