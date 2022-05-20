package hodometer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlattenStructToProerties(t *testing.T) {
	expected := properties{
		"cluster_id":               "cluster id",
		"seldon_core_version":      "0.1.2",
		"node_count":               uint(3),
		"model_count":              uint(4),
		"pipeline_count":           uint(5),
		"experiment_count":         uint(6),
		"server_count":             uint(7),
		"server_replica_count":     uint(8),
		"gpu_enabled_count":        uint(9),
		"server_cpu_cores_sum":     float32(10),
		"server_memory_gb_sum":     float32(11),
		"multimodel_enabled_count": uint(12),
		"overcommit_enabled_count": uint(13),
		"kubernetes_version":       "14.15.16",
		"is_global_installation":   false,
		"collector_version":        "17.18.19",
		"collector_git_commit":     "20",
	}

	m := UsageMetrics{}
	m.ClusterId = "cluster id"
	m.SeldonCoreVersion = "0.1.2"
	m.NodeCount = 3
	m.ModelCount = 4
	m.PipelineCount = 5
	m.ExperimentCount = 6
	m.ServerCount = 7
	m.ServerReplicaCount = 8
	m.GpuEnabledCount = 9
	m.ServerCpuCoresSum = 10.0
	m.ServerMemoryGbSum = 11.0
	m.MultimodelEnabledCount = 12
	m.OvercommitEnabledCount = 13
	m.KubernetesVersion = "14.15.16"
	m.IsGlobalInstallation = false
	m.CollectorVersion = "17.18.19"
	m.CollectorGitCommit = "20"

	type test struct {
		name    string
		metrics interface{}
	}

	tests := []test{
		{name: "raw struct", metrics: m},
		{name: "pointer to struct", metrics: &m},
	}

	for _, tt := range tests {
		actual := properties{}
		flattenStructToProperties(actual, tt.metrics)
		require.Equal(t, expected, actual)
	}
}
