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
