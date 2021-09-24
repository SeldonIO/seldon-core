package v1alpha1

type InferenceNextStep string

type InferenceStep struct {
	// Name of the step
	Name string `json:"name" protobuf:"bytes,1,name=name"`
	// Next synchronous step
	// +optional
	Next []InferenceNextStep `json:"next,omitempty" protobuf:"bytes,2,opt,name=next"`
	// Next asynchronous step
	// +optional
	NextAsync []InferenceNextStep `json:"nextAsync,omitempty" protobuf:"bytes,3,opt,name=nextAsync"`
	// Reference to model artifact
	// +optional
	Ref string `json:"ref,omitempty" protobuf:"bytes,4,opt,name=ref"`
	// Model artifact spec inline
	// +optional
	Spec InferenceArtifactSpec `json:"spec,omitempty" protobuf:"bytes,5,opt,name=spec"`
	// Condition specification to satisfy for this step to be run
	// +optional
	If string `json:"if,omitempty" protobuf:"bytes,6,opt,name=if"`
}
