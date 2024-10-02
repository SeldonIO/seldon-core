/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package experiment

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
)

func TestCreateExperiment(t *testing.T) {
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

func TestCreateExperimentFromSnapshot(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name     string
		proto    *scheduler.ExperimentSnapshot
		expected *Experiment
	}

	getStrPtr := func(val string) *string { return &val }
	tests := []test{
		{
			name: "experiment",
			proto: &scheduler.ExperimentSnapshot{
				Experiment: &scheduler.Experiment{
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
				Deleted: false,
			},
			expected: &Experiment{
				Name:         "foo",
				Active:       false,
				Deleted:      false,
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
			name: "deleted experiment",
			proto: &scheduler.ExperimentSnapshot{
				Experiment: &scheduler.Experiment{
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
				Deleted: true,
			},
			expected: &Experiment{
				Name:         "foo",
				Active:       false,
				Deleted:      true,
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			experiment := CreateExperimentFromSnapshot(test.proto)
			g.Expect(experiment).To(Equal(test.expected))
		})
	}

}
