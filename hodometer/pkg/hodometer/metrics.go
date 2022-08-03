package hodometer

type UsageMetrics struct {
	CollectorMetrics
	ClusterMetrics
	ResourceMetrics
	FeatureMetrics
}

type CollectorMetrics struct {
	CollectorVersion   string `json:"collector_version"`
	CollectorGitCommit string `json:"collector_git_commit"`
}

type ClusterMetrics struct {
	ClusterId         string `json:"cluster_id"`
	SeldonCoreVersion string `json:"seldon_core_version"`
	KubernetesMetrics
}

type KubernetesMetrics struct {
	KubernetesVersion string `json:"kubernetes_version"`
}

type ResourceMetrics struct {
	ModelCount         uint `json:"model_count"`
	PipelineCount      uint `json:"pipeline_count"`
	ExperimentCount    uint `json:"experiment_count"`
	ServerCount        uint `json:"server_count"`
	ServerReplicaCount uint `json:"server_replica_count"`
}

type FeatureMetrics struct {
	MultimodelEnabledCount uint    `json:"multimodel_enabled_count"`
	OvercommitEnabledCount uint    `json:"overcommit_enabled_count"`
	GpuEnabledCount        uint    `json:"gpu_enabled_count"`
	ServerCpuCoresSum      float32 `json:"server_cpu_cores_sum"`
	ServerMemoryGbSum      float32 `json:"server_memory_gb_sum"`
}
