package experiment

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	log "github.com/sirupsen/logrus"

	. "github.com/onsi/gomega"
)

func TestSaveAndRestore(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name        string
		experiments []*Experiment
	}

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
			for _, p := range test.experiments {
				g.Expect(cmp.Equal(p, es.experiments[p.Name])).To(BeTrue())
			}
		})
	}
}
