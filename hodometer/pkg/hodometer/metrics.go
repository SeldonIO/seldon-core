/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
