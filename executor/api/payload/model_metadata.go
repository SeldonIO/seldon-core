package payload

// Struct to store ModelMetadata deserialised from other payloads format (REST/GRPC).
// We need to deserialise this payload in order to build GraphMetadata.
// As deserialisation is REST/GRPC depenendent it must happens in corresponding client
// and ModelMetadata is returned by a corresponding method. It is therefore defined
// here and not in `predictor_process/metadata.go` in order to avoid circular dependency.
type ModelMetadata struct {
	Name     string            `json:"name,omitempty"`
	Platform string            `json:"platform,omitempty"`
	Versions []string          `json:"versions,omitempty"`
	Inputs   interface{}       `json:"inputs,omitempty"`
	Outputs  interface{}       `json:"outputs,omitempty"`
	Custom   map[string]string `json:"custom,omitempty"`
}
