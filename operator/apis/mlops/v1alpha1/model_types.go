/*
Copyright 2021.

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

package v1alpha1

import (
	"fmt"

	"knative.dev/pkg/apis"

	scheduler "github.com/seldonio/seldon-core/operatorv2/scheduler/apis/mlops/scheduler"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// ModelSpec defines the desired state of Model
type ModelSpec struct {
	InferenceArtifactSpec `json:",inline"`
	// List of extra requirements for this model to be loaded on a compatible server
	Requirements []string `json:"requirements,omitempty"`
	// Memory needed for model
	// +optional
	Memory *resource.Quantity `json:"memory,omitempty"`
	// Name of the Server to deploy this artifact
	// +optional
	Server *string `json:"server,omitempty"`
	// Number of replicas - default 1
	Replicas *int32 `json:"replicas,omitempty"`
	// Min number of replicas - default equal to replicas
	MinReplicas *int32 `json:"minReplicas,omitempty"`
	// Max number of replicas - default equal to replicas
	MaxReplicas *int32 `json:"maxReplicas,omitempty"`
	// Model already loaded on a server. Don't schedule.
	// Default false
	PreLoaded bool `json:"preloaded,omitempty"`
	// Dedicated server exclusive to this model
	// Default false
	Dedicated bool `json:"dedicated,omitempty"`
	// Payload logging
	Logger *LoggingSpec `json:"logger,omitempty"`
}

type LoggingSpec struct {
	//Percentage of payloads to log
	Percent *uint `json:"percent,omitempty"`
}

type InferenceArtifactSpec struct {
	// Model type
	// +optional
	ModelType *string `json:"modelType,omitempty"`
	// Storage URI for the model repository
	StorageURI string `json:"storageUri"`
	// Artifact Version
	// A v2 version folder to select from storage bucket
	// +optional
	ArtifactVersion *uint32 `json:"artifactVersion,omitempty"`
	// Schema URI
	// +optional
	SchemaURI *string `json:"schemaUri,omitempty"`
	// Secret name
	// +optional
	SecretName *string `json:"secretName,omitempty"`
}

// ModelStatus defines the observed state of Model
type ModelStatus struct {
	// Important: Run "make" to regenerate code after modifying this file
	duckv1.Status `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Model is the Schema for the models API
type Model struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ModelSpec   `json:"spec,omitempty"`
	Status ModelStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ModelList contains a list of Model
type ModelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Model `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Model{}, &ModelList{})
}

// Method to convert Model resource to scheduler proto for communication with Scheduler
func (m Model) AsSchedulerModel() (*scheduler.Model, error) {
	md := &scheduler.Model{
		Meta: &scheduler.MetaData{
			Name: m.Name,
			KubernetesMeta: &scheduler.KubernetesMeta{
				Namespace:  m.Namespace,
				Generation: m.Generation,
			},
		},
		ModelSpec: &scheduler.ModelSpec{
			Uri:          m.Spec.StorageURI,
			Requirements: m.Spec.Requirements,
			Server:       m.Spec.Server,
		},
		DeploymentSpec: &scheduler.DeploymentSpec{
			LogPayloads: m.Spec.Logger != nil, // Simple boolean switch at present
		},
	}
	// Add storage secret if specified
	if m.Spec.SecretName != nil {
		md.ModelSpec.StorageConfig = &scheduler.StorageConfig{
			Config: &scheduler.StorageConfig_StorageSecretName{
				StorageSecretName: *m.Spec.SecretName,
			},
		}
	}
	// Add modelType to requirements if specified
	if m.Spec.ModelType != nil {
		md.ModelSpec.Requirements = append(md.ModelSpec.Requirements, *m.Spec.ModelType)
	}
	// Set Replicas
	//TODO add min/max replicas
	if m.Spec.Replicas != nil {
		md.DeploymentSpec.Replicas = uint32(*m.Spec.Replicas)
	} else {
		md.DeploymentSpec.Replicas = 1
	}

	if m.Spec.MinReplicas != nil {
		md.DeploymentSpec.MinReplicas = uint32(*m.Spec.MinReplicas)
	} else {
		md.DeploymentSpec.MinReplicas = 1
	}

	if m.Spec.MaxReplicas != nil {
		md.DeploymentSpec.MaxReplicas = uint32(*m.Spec.MaxReplicas)
	}

	// Set memory bytes
	if m.Spec.Memory != nil {
		if i64, ok := m.Spec.Memory.AsInt64(); ok {
			ui64 := uint64(i64)
			md.ModelSpec.MemoryBytes = &ui64
		} else {
			return nil, fmt.Errorf("Can't convert model memory quantity to bytes. %s", m.Spec.Memory.String())
		}
	}
	return md, nil
}

const (
	DeploymentsReady apis.ConditionType = "DeploymentsReady"
	SeldonMeshReady  apis.ConditionType = "SeldonMeshReady"
)

var conditionSet = apis.NewLivingConditionSet(
	DeploymentsReady,
	SeldonMeshReady,
)

var _ apis.ConditionsAccessor = (*ModelStatus)(nil)

func (ms *ModelStatus) InitializeConditions() {
	conditionSet.Manage(ms).InitializeConditions()
}

func (ms *ModelStatus) IsReady() bool {
	return conditionSet.Manage(ms).IsHappy()
}

func (ms *ModelStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return conditionSet.Manage(ms).GetCondition(t)
}

func (ms *ModelStatus) IsConditionReady(t apis.ConditionType) bool {
	return conditionSet.Manage(ms).GetCondition(t) != nil && conditionSet.Manage(ms).GetCondition(t).Status == v1.ConditionTrue
}

func (ms *ModelStatus) SetCondition(conditionType apis.ConditionType, condition *apis.Condition) {
	switch {
	case condition == nil:
		conditionSet.Manage(ms).MarkUnknown(conditionType, "", "")
	case condition.Status == v1.ConditionUnknown:
		conditionSet.Manage(ms).MarkUnknown(conditionType, condition.Reason, condition.Message)
	case condition.Status == v1.ConditionTrue:
		conditionSet.Manage(ms).MarkTrueWithReason(conditionType, condition.Reason, condition.Message)
	case condition.Status == v1.ConditionFalse:
		conditionSet.Manage(ms).MarkFalse(conditionType, condition.Reason, condition.Message)
	}
}

func (ms *ModelStatus) CreateAndSetCondition(conditionType apis.ConditionType, isTrue bool, reason string) {
	condition := apis.Condition{}
	if isTrue {
		condition.Status = v1.ConditionTrue
	} else {
		condition.Status = v1.ConditionFalse
	}
	condition.Type = conditionType
	condition.Reason = reason
	condition.LastTransitionTime = apis.VolatileTime{
		Inner: metav1.Now(),
	}
	ms.SetCondition(conditionType, &condition)
}
