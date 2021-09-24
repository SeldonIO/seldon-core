package v1alpha1

type GraphExplainersSpec struct {
	// Start step name. If omitted will assume first step
	// +optional
	Start string `json:"start,omitempty" protobuf:"bytes,1,opt,name=start"`
	// End step name. If omitted will assume last step
	// +optional
	End string `json:"end,omitempty" protobuf:"bytes,2,opt,name=end"`
	// Specification of explainer
	Spec InferenceExplainerSpec `json:"spec" protobuf:"bytes,3,name=spec"`
}
