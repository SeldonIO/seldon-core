package v1alpha1

type InferenceNextStep string

type InferenceStep struct {
	Name      string              `json:"name" protobuf:"bytes,1,name=name"`
	Next      []InferenceNextStep `json:"next" protobuf:"bytes,2,opt,name=next"`
	NextAsync []InferenceNextStep `json:"nextAsync" protobuf:"bytes,2,opt,name=nextAsync"`
	Ref       string              `json:"ref" protobuf:"bytes,1,name=ref"`
	If        string              `json:"if,omitempty" protobuf:"bytes,1,opt,name=if"`
}
