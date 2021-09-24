package v1alpha1

type CustomServerSpec struct {
	Ref                     string `json:"ref,omitempty" protobuf:"bytes,1,opt,name=ref"`
	ServerCustomizationSpec `json:",inline"`
}
