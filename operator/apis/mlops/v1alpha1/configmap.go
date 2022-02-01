package v1alpha1

// +kubebuilder:object:generate=false
type SeldonConfig struct {
}

type ServerDefaults struct {
	Agent  AgentConfig  `json:"agent"`
	Rclone RcloneConfig `json:"rclone"`
}

type AgentConfig struct {
	Image           string `json:"image"`
	ImagePullPolicy string `json:"imagePullPolicy"`
}

type RcloneConfig struct {
	Image           string `json:"image"`
	ImagePullPolicy string `json:"imagePullPolicy"`
}
