package payload

type ModelMetadata struct {
	Name     string      `json:"name,omitempty"`
	Platform string      `json:"platform,omitempty"`
	Versions []string    `json:"versions,omitempty"`
	Inputs   interface{} `json:"inputs,omitempty"`
	Outputs  interface{} `json:"outputs,omitempty"`
}
