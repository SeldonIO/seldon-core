/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package experiment

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

func TestSaveAndRestore(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name        string
		experiments []*Experiment
		errors      []bool
	}

	getStrPtr := func(val string) *string { return &val }

	tests := []test{
		{
			name: "basic model experiment",
			experiments: []*Experiment{
				{
					Name: "test1",
					Candidates: []*Candidate{
						{
							Name:   "model1",
							Weight: 50,
						},
						{
							Name:   "model2",
							Weight: 50,
						},
					},
					Mirror: &Mirror{
						Name:    "model3",
						Percent: 90,
					},
					Config: &Config{
						StickySessions: true,
					},
					KubernetesMeta: &KubernetesMeta{
						Namespace:  "default",
						Generation: 2,
					},
				},
			},
			errors: []bool{false},
		},
		{
			name: "basic 2 model experiment",
			experiments: []*Experiment{
				{
					Name: "test1",
					Candidates: []*Candidate{
						{
							Name:   "model1",
							Weight: 50,
						},
						{
							Name:   "model2",
							Weight: 50,
						},
					},
					Mirror: &Mirror{
						Name:    "model3",
						Percent: 90,
					},
					Config: &Config{
						StickySessions: true,
					},
					KubernetesMeta: &KubernetesMeta{
						Namespace:  "default",
						Generation: 2,
					},
				},
				{
					Name: "test2",
					Candidates: []*Candidate{
						{
							Name:   "model1",
							Weight: 50,
						},
						{
							Name:   "model2",
							Weight: 50,
						},
					},
					Mirror: &Mirror{
						Name:    "model3",
						Percent: 90,
					},
					Config: &Config{
						StickySessions: true,
					},
					KubernetesMeta: &KubernetesMeta{
						Namespace:  "default",
						Generation: 2,
					},
				},
			},
			errors: []bool{false, false},
		},
		{
			name: "basic pipeline experiment",
			experiments: []*Experiment{
				{
					Name:         "test1",
					ResourceType: PipelineResourceType,
					Candidates: []*Candidate{
						{
							Name:   "pipeline1",
							Weight: 50,
						},
						{
							Name:   "pipeline2",
							Weight: 50,
						},
					},
					Mirror: &Mirror{
						Name:    "pipeline3",
						Percent: 90,
					},
					Config: &Config{
						StickySessions: true,
					},
					KubernetesMeta: &KubernetesMeta{
						Namespace:  "default",
						Generation: 2,
					},
				},
			},
			errors: []bool{false},
		},
		{
			name: "faulty experiment",
			experiments: []*Experiment{
				{
					Name:         "test1",
					ResourceType: ModelResourceType,
					Default:      getStrPtr("model1"),
					Candidates: []*Candidate{
						{
							Name:   "model1",
							Weight: 50,
						},
						{
							Name:   "model2",
							Weight: 50,
						},
					},
					KubernetesMeta: &KubernetesMeta{
						Namespace:  "default",
						Generation: 2,
					},
				},
				{ // duplicate default
					Name:         "test2",
					ResourceType: ModelResourceType,
					Default:      getStrPtr("model1"),
					Candidates: []*Candidate{
						{
							Name:   "model1",
							Weight: 50,
						},
						{
							Name:   "model3",
							Weight: 50,
						},
					},
					KubernetesMeta: &KubernetesMeta{
						Namespace:  "default",
						Generation: 2,
					},
				},
			},
			errors: []bool{false, true},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := fmt.Sprintf("%s/db", t.TempDir())
			logger := log.New()
			db, err := newExperimentDbManager(getExperimentDbFolder(path), logger)
			g.Expect(err).To(BeNil())
			for _, p := range test.experiments {
				err := db.save(p)
				g.Expect(err).To(BeNil())
			}
			err = db.Stop()
			g.Expect(err).To(BeNil())

			es := NewExperimentServer(log.New(), nil, nil, nil)
			err = es.InitialiseOrRestoreDB(path)
			g.Expect(err).To(BeNil())
			for idx, p := range test.experiments {
				if !test.errors[idx] {
					g.Expect(cmp.Equal(p, es.experiments[p.Name])).To(BeTrue())
				}
			}
		})
	}
}

func TestGetExperimentFromDB(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name           string
		experiments    []*Experiment
		experimentName string
		isErr          bool
	}

	tests := []test{
		{
			name: "basic 2 model experiment",
			experiments: []*Experiment{
				{
					Name: "test1",
					Candidates: []*Candidate{
						{
							Name:   "model1",
							Weight: 50,
						},
						{
							Name:   "model2",
							Weight: 50,
						},
					},
					Mirror: &Mirror{
						Name:    "model3",
						Percent: 90,
					},
					Config: &Config{
						StickySessions: true,
					},
					KubernetesMeta: &KubernetesMeta{
						Namespace:  "default",
						Generation: 2,
					},
				},
				{
					Name: "test2",
					Candidates: []*Candidate{
						{
							Name:   "model1",
							Weight: 50,
						},
						{
							Name:   "model2",
							Weight: 50,
						},
					},
					Mirror: &Mirror{
						Name:    "model3",
						Percent: 90,
					},
					Config: &Config{
						StickySessions: true,
					},
					KubernetesMeta: &KubernetesMeta{
						Namespace:  "default",
						Generation: 2,
					},
				},
			},
			experimentName: "test2",
			isErr:          false,
		},
		{
			name: "Experiment not found",
			experiments: []*Experiment{
				{
					Name:         "test1",
					ResourceType: ModelResourceType,
					Candidates: []*Candidate{
						{
							Name:   "model1",
							Weight: 50,
						},
						{
							Name:   "model2",
							Weight: 50,
						},
					},
					KubernetesMeta: &KubernetesMeta{
						Namespace:  "default",
						Generation: 2,
					},
				},
				{
					Name:         "test2",
					ResourceType: ModelResourceType,
					Candidates: []*Candidate{
						{
							Name:   "model1",
							Weight: 50,
						},
						{
							Name:   "model3",
							Weight: 50,
						},
					},
					KubernetesMeta: &KubernetesMeta{
						Namespace:  "default",
						Generation: 2,
					},
				},
			},
			experimentName: "test3",
			isErr:          true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := fmt.Sprintf("%s/db", t.TempDir())
			logger := log.New()
			db, err := newExperimentDbManager(getExperimentDbFolder(path), logger)
			g.Expect(err).To(BeNil())
			for _, p := range test.experiments {
				err := db.save(p)
				g.Expect(err).To(BeNil())
			}

			actualExperiment, err  := db.get(test.experimentName)
			if test.isErr {
				g.Expect(err).To(BeNil())
				g.Expect(actualExperiment).To(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(actualExperiment).ToNot(BeNil())
				expFound := false
				for _, exp := range test.experiments {
					if exp.Name == test.experimentName {
						g.Expect(cmp.Equal(exp, actualExperiment)).To(BeTrue())
						expFound = true
					}
				}
				g.Expect(expFound).To(BeTrue())
			}
			err = db.Stop()
			g.Expect(err).To(BeNil())
		})
	}
}
