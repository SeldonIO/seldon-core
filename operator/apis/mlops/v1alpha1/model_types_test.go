/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package v1alpha1

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	scheduler "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
)

func TestModelSpec_Validate(t *testing.T) {
	t.Parallel()

	g := NewGomegaWithT(t)
	tests := []struct {
		name    string
		spec    *ModelSpec
		wantErr bool
	}{
		{
			name: "Both explainer and llm set",
			spec: &ModelSpec{
				Explainer: &ExplainerSpec{},
				Llm:       &LlmSpec{},
			},
			wantErr: true,
		},
		{
			name: "Only explainer set",
			spec: &ModelSpec{
				Explainer: &ExplainerSpec{},
			},
			wantErr: false,
		},
		{
			name: "Only llm set",
			spec: &ModelSpec{
				Llm: &LlmSpec{},
			},
			wantErr: false,
		},
		{
			name:    "Neither explainer nor llm set",
			spec:    &ModelSpec{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.spec.Validate()

			if tt.wantErr {
				g.Expect(err).ToNot(BeNil())
				g.Expect(err.Error()).To(ContainSubstring("Can't have both explainer and llm in model spec."))
			} else {
				g.Expect(err).To(BeNil())
			}
		})
	}

}

func TestAsModelDetails(t *testing.T) {
	t.Parallel()

	g := NewGomegaWithT(t)
	type test struct {
		name    string
		model   *Model
		modelpb *scheduler.Model
		error   bool
	}

	secret := "secret"
	modelType := "sklearn"
	server := "server"
	m1 := resource.MustParse("1M")
	m1bytes := uint64(1_000_000)
	incomeModel, llmModel := "income", "chat-gpt"
	tests := []test{
		{
			name: "simple",
			model: &Model{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "foo",
					Namespace:       "default",
					ResourceVersion: "1",
					Generation:      1,
				},
				Spec: ModelSpec{
					InferenceArtifactSpec: InferenceArtifactSpec{
						StorageURI: "gs://test",
					},
				},
			},
			modelpb: &scheduler.Model{
				Meta: &scheduler.MetaData{
					Name: "foo",
					KubernetesMeta: &scheduler.KubernetesMeta{
						Namespace:  "default",
						Generation: 1,
					},
				},
				ModelSpec: &scheduler.ModelSpec{
					Uri: "gs://test",
				},
				DeploymentSpec: &scheduler.DeploymentSpec{
					Replicas:    1,
					MinReplicas: 0,
					MaxReplicas: 0,
				},
			},
		},
		{
			name: "complex - explainer",
			model: &Model{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "foo",
					Namespace:  "default",
					Generation: 1,
				},
				Spec: ModelSpec{
					InferenceArtifactSpec: InferenceArtifactSpec{
						ModelType:  &modelType,
						StorageURI: "gs://test",
						SecretName: &secret,
					},
					Logger:       &LoggingSpec{},
					Requirements: []string{"a", "b"},
					ScalingSpec:  ScalingSpec{Replicas: i32(4)},
					Server:       &server,
					Explainer: &ExplainerSpec{
						Type:     "anchor_tabular",
						ModelRef: &incomeModel,
					},
					Parameters: []ParameterSpec{
						{
							Name:  "foo",
							Value: "bar",
						},
						{
							Name:  "foo2",
							Value: "bar2",
						},
					},
				},
			},
			modelpb: &scheduler.Model{
				Meta: &scheduler.MetaData{
					Name: "foo",
					KubernetesMeta: &scheduler.KubernetesMeta{
						Namespace:  "default",
						Generation: 1,
					},
				},
				ModelSpec: &scheduler.ModelSpec{
					Uri:           "gs://test",
					Requirements:  []string{"a", "b", modelType},
					StorageConfig: &scheduler.StorageConfig{Config: &scheduler.StorageConfig_StorageSecretName{StorageSecretName: secret}},
					Server:        &server,
					ModelSpec: &scheduler.ModelSpec_Explainer{
						Explainer: &scheduler.ExplainerSpec{
							Type:     "anchor_tabular",
							ModelRef: &incomeModel,
						},
					},
					Parameters: []*scheduler.ParameterSpec{
						{
							Name:  "foo",
							Value: "bar",
						},
						{
							Name:  "foo2",
							Value: "bar2",
						},
					},
				},
				DeploymentSpec: &scheduler.DeploymentSpec{
					Replicas:    4,
					LogPayloads: true,
					MinReplicas: 0,
					MaxReplicas: 0,
				},
			},
		},
		{
			name: "complex - llm",
			model: &Model{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "foo",
					Namespace:  "default",
					Generation: 1,
				},
				Spec: ModelSpec{
					InferenceArtifactSpec: InferenceArtifactSpec{
						ModelType:  &modelType,
						StorageURI: "gs://test",
						SecretName: &secret,
					},
					Logger:       &LoggingSpec{},
					Requirements: []string{"a", "b"},
					ScalingSpec:  ScalingSpec{Replicas: i32(4)},
					Server:       &server,
					Llm: &LlmSpec{
						ModelRef: &llmModel,
					},
					Parameters: []ParameterSpec{
						{
							Name:  "foo",
							Value: "bar",
						},
						{
							Name:  "foo2",
							Value: "bar2",
						},
					},
				},
			},
			modelpb: &scheduler.Model{
				Meta: &scheduler.MetaData{
					Name: "foo",
					KubernetesMeta: &scheduler.KubernetesMeta{
						Namespace:  "default",
						Generation: 1,
					},
				},
				ModelSpec: &scheduler.ModelSpec{
					Uri:           "gs://test",
					Requirements:  []string{"a", "b", modelType},
					StorageConfig: &scheduler.StorageConfig{Config: &scheduler.StorageConfig_StorageSecretName{StorageSecretName: secret}},
					Server:        &server,
					ModelSpec: &scheduler.ModelSpec_Llm{
						Llm: &scheduler.LlmSpec{
							ModelRef: &llmModel,
						},
					},
					Parameters: []*scheduler.ParameterSpec{
						{
							Name:  "foo",
							Value: "bar",
						},
						{
							Name:  "foo2",
							Value: "bar2",
						},
					},
				},
				DeploymentSpec: &scheduler.DeploymentSpec{
					Replicas:    4,
					LogPayloads: true,
					MinReplicas: 0,
					MaxReplicas: 0,
				},
			},
		},
		{
			name: "memory",
			model: &Model{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "foo",
					Namespace:  "default",
					Generation: 1,
				},
				Spec: ModelSpec{
					InferenceArtifactSpec: InferenceArtifactSpec{
						StorageURI: "gs://test",
					},
					Memory:      &m1,
					ScalingSpec: ScalingSpec{Replicas: i32(1)},
				},
			},
			modelpb: &scheduler.Model{
				Meta: &scheduler.MetaData{
					Name: "foo",
					KubernetesMeta: &scheduler.KubernetesMeta{
						Namespace:  "default",
						Generation: 1,
					},
				},
				ModelSpec: &scheduler.ModelSpec{
					Uri:         "gs://test",
					MemoryBytes: &m1bytes,
				},
				DeploymentSpec: &scheduler.DeploymentSpec{
					Replicas:    1,
					MinReplicas: 0,
					MaxReplicas: 0,
				},
			},
		},
		{
			name: "simple min replica",
			model: &Model{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "foo",
					Namespace:       "default",
					ResourceVersion: "1",
					Generation:      1,
				},
				Spec: ModelSpec{
					InferenceArtifactSpec: InferenceArtifactSpec{
						StorageURI: "gs://test",
					},
					ScalingSpec: ScalingSpec{MinReplicas: i32(4)},
				},
			},
			modelpb: &scheduler.Model{
				Meta: &scheduler.MetaData{
					Name: "foo",
					KubernetesMeta: &scheduler.KubernetesMeta{
						Namespace:  "default",
						Generation: 1,
					},
				},
				ModelSpec: &scheduler.ModelSpec{
					Uri: "gs://test",
				},
				DeploymentSpec: &scheduler.DeploymentSpec{
					Replicas:    4,
					MinReplicas: 4,
					MaxReplicas: 0,
				},
			},
		},
		{
			name: "simple max replica",
			model: &Model{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "foo",
					Namespace:       "default",
					ResourceVersion: "1",
					Generation:      1,
				},
				Spec: ModelSpec{
					InferenceArtifactSpec: InferenceArtifactSpec{
						StorageURI: "gs://test",
					},
					ScalingSpec: ScalingSpec{Replicas: i32(1), MaxReplicas: i32(4)},
				},
			},
			modelpb: &scheduler.Model{
				Meta: &scheduler.MetaData{
					Name: "foo",
					KubernetesMeta: &scheduler.KubernetesMeta{
						Namespace:  "default",
						Generation: 1,
					},
				},
				ModelSpec: &scheduler.ModelSpec{
					Uri: "gs://test",
				},
				DeploymentSpec: &scheduler.DeploymentSpec{
					Replicas:    1,
					MinReplicas: 0,
					MaxReplicas: 4,
				},
			},
		},
		{
			name: "range violation min",
			model: &Model{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "foo",
					Namespace:       "default",
					ResourceVersion: "1",
					Generation:      1,
				},
				Spec: ModelSpec{
					InferenceArtifactSpec: InferenceArtifactSpec{
						StorageURI: "gs://test",
					},
					ScalingSpec: ScalingSpec{MinReplicas: i32(4), Replicas: i32(1)},
				},
			},
			modelpb: &scheduler.Model{
				Meta: &scheduler.MetaData{
					Name: "foo",
					KubernetesMeta: &scheduler.KubernetesMeta{
						Namespace:  "default",
						Generation: 1,
					},
				},
				ModelSpec: &scheduler.ModelSpec{
					Uri: "gs://test",
				},
				DeploymentSpec: &scheduler.DeploymentSpec{
					Replicas:    1,
					MinReplicas: 0,
					MaxReplicas: 4,
				},
			},
			error: true,
		},
		{
			name: "range violation max",
			model: &Model{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "foo",
					Namespace:       "default",
					ResourceVersion: "1",
					Generation:      1,
				},
				Spec: ModelSpec{
					InferenceArtifactSpec: InferenceArtifactSpec{
						StorageURI: "gs://test",
					},
					ScalingSpec: ScalingSpec{Replicas: i32(4), MaxReplicas: i32(1)},
				},
			},
			modelpb: &scheduler.Model{
				Meta: &scheduler.MetaData{
					Name: "foo",
					KubernetesMeta: &scheduler.KubernetesMeta{
						Namespace:  "default",
						Generation: 1,
					},
				},
				ModelSpec: &scheduler.ModelSpec{
					Uri: "gs://test",
				},
				DeploymentSpec: &scheduler.DeploymentSpec{
					Replicas:    1,
					MinReplicas: 0,
					MaxReplicas: 4,
				},
			},
			error: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			md, err := test.model.AsSchedulerModel()
			if !test.error {
				g.Expect(err).To(BeNil())
				g.Expect(md).To(Equal(test.modelpb))
			} else {
				g.Expect(err).ToNot(BeNil())
			}
		})
	}
}

func i32(i int32) *int32 { return &i }

func TestGetValidatedScalingSpec(t *testing.T) {
	type test struct {
		name        string
		replicas    *int32
		minReplicas *int32
		maxReplicas *int32
		expected    *ValidatedScalingSpec
		wantErr     string
	}

	g := NewGomegaWithT(t)

	tests := []test{
		{
			name:        "happy path replicas",
			replicas:    i32(2),
			minReplicas: i32(1),
			maxReplicas: i32(3),
			expected: &ValidatedScalingSpec{
				Replicas:    2,
				MinReplicas: 1,
				MaxReplicas: 3,
			},
			wantErr: "",
		},
		{
			name:        "replicas is less than min replicas",
			replicas:    i32(1),
			minReplicas: i32(2),
			maxReplicas: i32(4),
			expected:    nil,
			wantErr:     "number of replicas 1 must be >= min replicas  2",
		},
		{
			name:        "replicas is bigger than max replicas",
			replicas:    i32(5),
			minReplicas: i32(1),
			maxReplicas: i32(4),
			expected:    nil,
			wantErr:     "number of replicas 5 must be <= max replicas  4",
		},
		{
			name:        "replicas is smaller than min replicas adjusted to min replicas",
			replicas:    i32(0),
			minReplicas: i32(1),
			maxReplicas: i32(4),
			expected: &ValidatedScalingSpec{
				Replicas:    1,
				MinReplicas: 1,
				MaxReplicas: 4,
			},
			wantErr: "",
		},
		{
			name:        "replicas stays at 0 when min replicas and max replicas is 4",
			replicas:    i32(0),
			minReplicas: i32(0),
			maxReplicas: i32(4),
			expected: &ValidatedScalingSpec{
				Replicas:    0,
				MinReplicas: 0,
				MaxReplicas: 4,
			},
			wantErr: "",
		},
		{
			name:        "replicas gets adjusted to min replicas when is less than min replicas",
			replicas:    i32(0),
			minReplicas: i32(2),
			maxReplicas: i32(4),
			expected: &ValidatedScalingSpec{
				Replicas:    2,
				MinReplicas: 2,
				MaxReplicas: 4,
			},
			wantErr: "",
		},
		{
			name:        "no auto scaling enabled",
			replicas:    i32(2),
			minReplicas: nil,
			maxReplicas: nil,
			expected: &ValidatedScalingSpec{
				Replicas: 2,
			},
			wantErr: "",
		},
		{
			name:        "unset replica params defaults to 1",
			replicas:    nil,
			minReplicas: nil,
			maxReplicas: nil,
			expected: &ValidatedScalingSpec{
				Replicas:    1,
				MinReplicas: 0,
				MaxReplicas: 0,
			},
			wantErr: "",
		},
		{
			name:        "unset replica params defaults to 1",
			replicas:    nil,
			minReplicas: i32(1),
			maxReplicas: nil,
			expected: &ValidatedScalingSpec{
				Replicas:    1,
				MinReplicas: 1,
				MaxReplicas: 0,
			},
			wantErr: "",
		},
		{
			name:        "min replica to 1 and replica to 0 should convert to min replica",
			replicas:    i32(0),
			minReplicas: i32(1),
			maxReplicas: nil,
			expected: &ValidatedScalingSpec{
				Replicas:    1,
				MinReplicas: 1,
				MaxReplicas: 0,
			},
			wantErr: "",
		},
		{
			name:        "min replicas scaling",
			replicas:    nil,
			minReplicas: i32(2),
			maxReplicas: nil,
			expected: &ValidatedScalingSpec{
				Replicas:    2,
				MinReplicas: 2,
			},
			wantErr: "",
		},
		{
			name:        "max replicas scaling",
			replicas:    nil,
			minReplicas: nil,
			maxReplicas: i32(2),
			expected: &ValidatedScalingSpec{
				Replicas:    1,
				MinReplicas: 0,
				MaxReplicas: 2,
			},
			wantErr: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scalingSpec, err := GetValidatedScalingSpec(test.replicas, test.minReplicas, test.maxReplicas)

			if test.wantErr != "" {
				if err == nil {
					t.Errorf("expected error: %v, got nil", test.wantErr)
					return
				}
				if err.Error() != test.wantErr {
					t.Errorf("expected error: %q, got: %q", test.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			g.Expect(scalingSpec).To(Equal(test.expected))
		})
	}

}

/* WARNING: Read this first if test below fails (either at compile-time or while running the
* test):
*
* The test below aims to ensure that the fields used in kubebuilder:printcolumn comments in
* model_types.go match the structure and condition types in the Model CRD.
*
* If the test fails, it means that the CRD was updated without updating the kubebuilder:
* printcolumn comments.
*
* Rather than fixing the test directly, FIRST UPDATE the kubebuilder:printcolumn comments,
* THEN update the test to also match the new values.
*
 */
func TestModelStatusPrintColumns(t *testing.T) {
	t.Parallel()

	g := NewGomegaWithT(t)

	type test struct {
		name                          string
		model                         *Model
		expectedJsonSerializationKeys []string
	}
	condition_fail_reason := "Model.Status.Conditions[].Type string changed in CRD from \"%s\" to \"%[2]s\" without updating kubebuilder:printcolumn comment for type Model. " +
		"Update kubebuilder:printcolumn comment in model_types.go to match \"%[2]s\"."
	json_fail_reason := "Json serialization of Model.Status does not contain path \"%s\", used in kubebuilder.printcolum comments. " +
		"Update kubebuilder:printcolumn comments to match the new serialization keys."

	// !! When the test fails, update the string values here ONLY after updating the kubebuilder:
	// printcolumn comments in model_types.go to match the new values
	//
	// The key represents the condition used by the CR, the value is the string currently used in
	// the kubebuilder:printcolumn comments.
	expectedPrintColumnString := map[apis.ConditionType]string{
		ModelReady:          "ModelReady",
		apis.ConditionReady: "Ready",
	}

	tests := []test{
		{
			name: "model ready conditions",
			model: &Model{
				Status: ModelStatus{
					Status: duckv1.Status{
						Conditions: duckv1.Conditions{
							// pass the strings used in kubebuilder:printcolumn comments as the ConditionType
							{Type: apis.ConditionType(expectedPrintColumnString[ModelReady]), Status: v1.ConditionTrue},
							{Type: apis.ConditionType(expectedPrintColumnString[apis.ConditionReady]), Status: v1.ConditionTrue},
						},
					},
				},
			},
		},
		{
			name: "model replicas",
			model: &Model{
				Status: ModelStatus{
					Replicas: 1,
				},
			},
			expectedJsonSerializationKeys: []string{"status.replicas"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if test.model.Status.Conditions != nil {
				searchMap := make(map[string]v1.ConditionStatus)
				for _, cond := range test.model.Status.Conditions {
					searchMap[string(cond.Type)] = cond.Status
				}
				for conditionKey, printColumnString := range expectedPrintColumnString {
					_, found := searchMap[string(conditionKey)]
					g.Expect(found).To(BeTrueBecause(condition_fail_reason, printColumnString, string(conditionKey)))
				}
			} else {
				jsonBytes, err := json.Marshal(test.model)
				g.Expect(err).To(BeNil())
				for _, key := range test.expectedJsonSerializationKeys {
					g.Expect(gjson.GetBytes(jsonBytes, key).Exists()).To(BeTrueBecause(json_fail_reason, key))
				}
			}
		})
	}
}
