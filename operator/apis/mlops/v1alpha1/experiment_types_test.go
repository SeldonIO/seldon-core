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

	"knative.dev/pkg/ptr"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"

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

/* WARNING: Read this first if test below fails (either at compile-time or while running the
* test):
*
* The test below aims to ensure that the fields used in kubebuilder:printcolumn comments in
* experiment_types.go match the structure and condition types in the Experiment CRD.
*
* If the test fails, it means that the CRD was updated without updating the kubebuilder:
* printcolumn comments.
*
* Rather than fixing the test directly, FIRST UPDATE the kubebuilder:printcolumn comments,
* THEN update the test to also match the new values.
*
 */
func TestExperimentStatusPrintColumns(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name       string
		experiment *Experiment
	}
	fail_reason := "Experiment.Status.Conditions[].Type string changed in CRD from \"%s\" to \"%[2]s\" without updating kubebuilder:printcolumn comment for type Experiment. " +
		"Update kubebuilder:printcolumn comment in experiment_types.go to match \"%[2]s\"."

	// !! When the test fails, update the string values here ONLY after updating the kubebuilder:
	// printcolumn comments in experiment_types.go to match the new values
	//
	// The key represents the condition used by the CR, the value is the string currently used in
	// the kubebuilder:printcolumn comments.
	expectedPrintColumnString := map[apis.ConditionType]string{
		ExperimentReady:     "ExperimentReady",
		CandidatesReady:     "CandidatesReady",
		MirrorReady:         "MirrorReady",
		apis.ConditionReady: "Ready",
	}

	tests := []test{
		{
			name: "experiment ready conditions",
			experiment: &Experiment{
				Status: ExperimentStatus{
					Status: duckv1.Status{
						Conditions: duckv1.Conditions{
							// pass the strings used in kubebuilder:printcolumn comments as the ConditionType
							{Type: apis.ConditionType(expectedPrintColumnString[ExperimentReady]), Status: v1.ConditionTrue},
							{Type: apis.ConditionType(expectedPrintColumnString[CandidatesReady]), Status: v1.ConditionTrue},
							{Type: apis.ConditionType(expectedPrintColumnString[MirrorReady]), Status: v1.ConditionTrue},
							{Type: apis.ConditionType(expectedPrintColumnString[apis.ConditionReady]), Status: v1.ConditionTrue},
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.experiment.Status.Conditions != nil {
				searchMap := make(map[string]v1.ConditionStatus)
				for _, cond := range test.experiment.Status.Conditions {
					searchMap[string(cond.Type)] = cond.Status
				}
				for conditionKey, printColumnString := range expectedPrintColumnString {
					_, found := searchMap[string(conditionKey)]
					g.Expect(found).To(BeTrueBecause(fail_reason, printColumnString, string(conditionKey)))
				}
			}
		})
	}
}

func TestExperimentStatusSetCondition(t *testing.T) {
	type args struct {
		conditionType apis.ConditionType
		condition     *apis.Condition
	}
	tests := []struct {
		name string
		args args
		want *v1.ConditionStatus
	}{
		{
			name: "should not panic if condition is nil",
			args: args{
				conditionType: ExperimentReady,
				condition:     nil,
			},
			want: nil,
		},
		{
			name: "ConditionUnknown",
			args: args{
				conditionType: ExperimentReady,
				condition: &apis.Condition{
					Status: "Unknown",
				},
			},
			want: (*v1.ConditionStatus)(ptr.String("Unknown")),
		},
		{
			name: "ConditionTrue",
			args: args{
				conditionType: ExperimentReady,
				condition: &apis.Condition{
					Status: "True",
				},
			},
			want: (*v1.ConditionStatus)(ptr.String("True")),
		},
		{
			name: "ConditionFalse",
			args: args{
				conditionType: ExperimentReady,
				condition: &apis.Condition{
					Status: "False",
				},
			},
			want: (*v1.ConditionStatus)(ptr.String("False")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			es := &ExperimentStatus{}
			es.SetCondition(tt.args.conditionType, tt.args.condition)
			if tt.want == nil {
				if es.GetCondition(tt.args.conditionType) != nil {
					t.Errorf("want %v : got %v", tt.want, es.GetCondition(tt.args.conditionType))
				}
			}
			if tt.want != nil {
				got := es.GetCondition(tt.args.conditionType).Status
				if *tt.want != got {
					t.Errorf("want %v : got %v", *tt.want, got)
				}
			}
		})
	}
}
