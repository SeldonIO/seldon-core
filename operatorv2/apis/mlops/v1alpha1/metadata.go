package v1alpha1

type ArtifactMetadataSpec struct {
	// swagger URI endpoint
	// +optional
	SwaggerURI *string `json:"swaggerURI,omitempty" protobuf:"bytes,9,opt,name=swaggerURI"`
	// model server setting URI endpoint
	// +optional
	ModelSettingsURI *string `json:"modelSettingsURI,omitempy" protobuf:"bytes,10,opt,name=modelSettingsURI"`
	// Prediction schema URI endpoint
	// +optional
	PredictionSchemaURI *string `json:"predictionSchemaURI,omitempy" protobuf:"bytes,11,opt,name=predictionSchemaURI"`
}