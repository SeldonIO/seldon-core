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
	duckv1 "knative.dev/pkg/apis/duck/v1"
	v1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"

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
							Name:   "b",
							Inputs: []string{"a"},
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
						Name:   "b",
						Inputs: []string{"a"},
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

/* WARNING: Read this first if test below fails (either at compile-time or while running the
* test):
*
* The test below aims to ensure that the fields used in kubebuilder:printcolumn comments in
* pipeline_types.go match the structure and condition types in the Pipeline CRD.
*
* If the test fails, it means that the CRD was updated without updating the kubebuilder:
* printcolumn comments.
*
* Rather than fixing the test directly, FIRST UPDATE the kubebuilder:printcolumn comments,
* THEN update the test to also match the new values.
*
*/
func TestPipelineStatusPrintColumns(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name			string
		pipeline	*Pipeline
	}
	fail_reason := "Pipeline.Status.Conditions[].Type string changed in CRD from \"%s\" to \"%[2]s\" without updating kubebuilder:printcolumn comment for type Model. " +
								 "Update kubebuilder:printcolumn comment in pipeline_types.go to match \"%[2]s\"."

	// !! When the test fails, update the string values here ONLY after updating the kubebuilder:
	// printcolumn comments in pipeline_types.go to match the new values
	//
	// The key represents the condition used by the CR, the value is the string currently used in
	// the kubebuilder:printcolumn comments.
	expectedPrintColumnString := map[apis.ConditionType]string{
		ModelsReady: "ModelsReady",
		PipelineReady: "PipelineReady",
		apis.ConditionReady: "Ready",
	}


	tests := []test{
		{
			name: "pipeline ready conditions",
			pipeline: &Pipeline{
				Status: PipelineStatus{
					Status: duckv1.Status{
						Conditions: duckv1.Conditions{
							// pass the strings used in kubebuilder:printcolumn comments as the ConditionType
							{ Type: apis.ConditionType(expectedPrintColumnString[ModelsReady]), Status: v1.ConditionTrue },
							{ Type: apis.ConditionType(expectedPrintColumnString[PipelineReady]), Status: v1.ConditionTrue },
							{ Type: apis.ConditionType(expectedPrintColumnString[apis.ConditionReady]), Status: v1.ConditionTrue },
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.pipeline.Status.Conditions != nil {
				searchMap := make (map[string]v1.ConditionStatus)
				for _, cond := range test.pipeline.Status.Conditions {
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
