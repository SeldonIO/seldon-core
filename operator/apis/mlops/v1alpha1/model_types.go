/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package v1alpha1

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	scheduler "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
)

// ModelSpec defines the desired state of Model
type ModelSpec struct {
	InferenceArtifactSpec `json:",inline"`
	// List of extra requirements for this model to be loaded on a compatible server
	Requirements []string `json:"requirements,omitempty"`
	// Memory needed for model
	// +optional
	Memory *resource.Quantity `json:"memory,omitempty"`
	// Scaling spec
	ScalingSpec `json:",inline"`
	// Name of the Server to deploy this artifact
	// +optional
	Server *string `json:"server,omitempty"`
	// Model already loaded on a server. Don't schedule.
	// Default false
	PreLoaded bool `json:"preloaded,omitempty"`
	// Dedicated server exclusive to this model
	// Default false
	Dedicated bool `json:"dedicated,omitempty"`
	// Payload logging
	Logger *LoggingSpec `json:"logger,omitempty"`
	// Explainer spec
	Explainer *ExplainerSpec `json:"explainer,omitempty"`
	// Parameters to load with model
	Parameters []ParameterSpec `json:"parameters,omitempty"`
}

type ParameterSpec struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Either ModelRef or PipelineRef is required
type ExplainerSpec struct {
	// type of explainer
	Type string `json:"type,omitempty"`
	// one of the following need to be set for blackbox explainers
	// Reference to Model
	// +optional
	ModelRef *string `json:"modelRef,omitempty"`
	// Reference to Pipeline
	// +optional
	PipelineRef *string `json:"pipelineRef,omitempty"`
}

type ScalingSpec struct {
	// Number of replicas - default 1
	Replicas *int32 `json:"replicas,omitempty"`
	// Min number of replicas - default equal to 0
	MinReplicas *int32 `json:"minReplicas,omitempty"`
	// Max number of replicas - default equal to 0
	MaxReplicas *int32 `json:"maxReplicas,omitempty"`
}

func (s *ScalingSpec) Default() {
	defaultReplicas := int32(1)
	if s.Replicas == nil {
		s.Replicas = &defaultReplicas
	}
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
	// Total number of replicas targeted by this model
	Replicas int32 `json:"replicas,omitempty"`
	// Number of available replicas
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`
	duckv1.Status     `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=mlm
//+kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas
//+kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="ModelReady")].status`,description="Model ready status"
//+kubebuilder:printcolumn:name="Desired Replicas",type=integer,JSONPath=`.spec.replicas`,description="Number of desired replicas"
//+kubebuilder:printcolumn:name="Available Replicas",type=integer,JSONPath=`.status.availableReplicas`,description="Number of replicas available to receive inference requests"
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

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
			Uri:             m.Spec.StorageURI,
			ArtifactVersion: m.Spec.ArtifactVersion,
			Requirements:    m.Spec.Requirements,
			Server:          m.Spec.Server,
		},
		DeploymentSpec: &scheduler.DeploymentSpec{
			LogPayloads: m.Spec.Logger != nil, // Simple boolean switch at present
		},
	}
	if m.Spec.Explainer != nil {
		md.ModelSpec.Explainer = &scheduler.ExplainerSpec{
			Type:        m.Spec.Explainer.Type,
			ModelRef:    m.Spec.Explainer.ModelRef,
			PipelineRef: m.Spec.Explainer.PipelineRef,
		}
	}
	if len(m.Spec.Parameters) > 0 {
		var parameters []*scheduler.ParameterSpec
		for _, param := range m.Spec.Parameters {
			parameters = append(parameters, &scheduler.ParameterSpec{
				Name:  param.Name,
				Value: param.Value,
			})
		}
		md.ModelSpec.Parameters = parameters
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
	if m.Spec.Replicas != nil {
		md.DeploymentSpec.Replicas = uint32(*m.Spec.Replicas)
	} else {
		if m.Spec.MinReplicas != nil {
			// set replicas to the min replicas if not set
			md.DeploymentSpec.Replicas = uint32(*m.Spec.MinReplicas)
		} else {
			md.DeploymentSpec.Replicas = 1
		}
	}

	if m.Spec.MinReplicas != nil {
		md.DeploymentSpec.MinReplicas = uint32(*m.Spec.MinReplicas)
		if md.DeploymentSpec.Replicas < md.DeploymentSpec.MinReplicas {
			return nil, fmt.Errorf("Number of replicas %d should be >= min replicas %d", md.DeploymentSpec.Replicas, md.DeploymentSpec.MinReplicas)
		}
	} else {
		md.DeploymentSpec.MinReplicas = 0
	}

	if m.Spec.MaxReplicas != nil {
		md.DeploymentSpec.MaxReplicas = uint32(*m.Spec.MaxReplicas)
		if md.DeploymentSpec.Replicas > md.DeploymentSpec.MaxReplicas {
			return nil, fmt.Errorf("Number of replicas %d should be <= max replicas %d", md.DeploymentSpec.Replicas, md.DeploymentSpec.MaxReplicas)
		}
	} else {
		md.DeploymentSpec.MaxReplicas = 0
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
