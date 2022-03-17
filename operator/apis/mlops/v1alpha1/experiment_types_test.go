package v1alpha1

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/operatorv2/scheduler/apis/mlops/scheduler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			name: "basic",
			experiment: &Experiment{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "foo",
					Namespace:  "default",
					Generation: 1,
				},
				Spec: ExperimentSpec{
					DefaultModel: getStrPtr("model1"),
					Candidates: []ExperimentCandidate{
						{
							ModelName: "model1",
							Weight:    20,
						},
						{
							ModelName: "model2",
							Weight:    30,
						},
					},
					Mirror: &ExperimentMirror{
						ModelName: "model4",
						Percent:   40,
					},
				},
			},
			proto: &scheduler.Experiment{
				Name:         "foo",
				DefaultModel: getStrPtr("model1"),
				Candidates: []*scheduler.ExperimentCandidate{
					{
						ModelName: "model1",
						Weight:    20,
					},
					{
						ModelName: "model2",
						Weight:    30,
					},
				},
				Mirror: &scheduler.ExperimentMirror{
					ModelName: "model4",
					Percent:   40,
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
