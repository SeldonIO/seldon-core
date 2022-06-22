package processor

import (
	"io/ioutil"
	"strconv"
	"testing"
	"time"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	"github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	"github.com/seldonio/seldon-core/scheduler/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/xdscache"
	"github.com/seldonio/seldon-core/scheduler/pkg/store"
	"github.com/seldonio/seldon-core/scheduler/pkg/store/experiment"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func addServer(
	ip *IncrementalProcessor,
	serverName string,
	numReplicas int,
	b *testing.B,
) {
	for replicaIdx := 0; replicaIdx < numReplicas; replicaIdx++ {
		err := ip.modelStore.AddServerReplica(
			&agent.AgentSubscribeRequest{
				ServerName:           serverName,
				Shared:               true,
				ReplicaIdx:           uint32(replicaIdx),
				ReplicaConfig:        nil,
				LoadedModels:         nil,
				AvailableMemoryBytes: 1_000,
			},
		)
		require.NoError(b, err)
	}
}

func addModel(
	ip *IncrementalProcessor,
	modelName string,
	modelVersion int,
	serverName string,
	b *testing.B,
) {
	// Load model
	model := &scheduler.Model{
		Meta: &scheduler.MetaData{
			Name: modelName,
		},
		ModelSpec:      &scheduler.ModelSpec{},
		DeploymentSpec: &scheduler.DeploymentSpec{},
	}

	err := ip.modelStore.UpdateModel(&scheduler.LoadModelRequest{Model: model})
	require.NoError(b, err)

	// Schedule model
	server, err := ip.modelStore.GetServer(serverName, true, false)
	require.NoError(b, err)

	replicas := []*store.ServerReplica{}
	replicaStatuses := make(map[int]store.ReplicaStatus)
	for i, r := range server.Replicas {
		replicas = append(replicas, r)
		replicaStatuses[i] = store.ReplicaStatus{State: store.Available}
	}

	err = ip.modelStore.UpdateLoadedModels(
		modelName,
		uint32(modelVersion),
		server.Name,
		replicas,
	)
	require.NoError(b, err)

	// Load model on agent(s)
	for replicaIdx := range replicas {
		err = ip.modelStore.UpdateModelState(
			modelName,
			uint32(modelVersion),
			server.Name,
			replicaIdx,
			nil,
			store.LoadRequested,
			store.Loaded,
			"",
		)
		require.NoError(b, err)
	}
}

func benchmarkModelUpdate(
	b *testing.B,
	numModels int,
	updatesPerModel int,
	numServerReplicas int,
	batchWaitMillis int,
) {
	const (
		serverName = "server1"
	)

	for i := 0; i < b.N; i++ {
		logger := logrus.New()
		logger.Out = ioutil.Discard

		eventHub, err := coordinator.NewEventHub(logger)
		require.NoError(b, err)

		memoryStore := store.NewMemoryStore(logger, store.NewLocalSchedulerStore(), eventHub)

		ip := NewIncrementalProcessor(
			cache.NewSnapshotCache(false, cache.IDHash{}, logger),
			"some node",
			logger,
			memoryStore,
			experiment.NewExperimentServer(logger, nil, memoryStore),
			nil,
			eventHub,
			&xdscache.PipelineGatewayDetails{
				Host:     "some host",
				HttpPort: 1,
				GrpcPort: 2,
			},
		)

		ip.batchWaitMillis = time.Duration(batchWaitMillis) * time.Millisecond

		addServer(ip, serverName, numServerReplicas, b)

		for modelVersion := 0; modelVersion < updatesPerModel; modelVersion++ {
			for modelId := 0; modelId < numModels; modelId++ {
				modelName := "model" + strconv.Itoa(modelId)

				addModel(ip, modelName, modelVersion+1, serverName, b)
			}
		}

		for len(ip.pendingModelVersions) > 0 {
			<-time.After(time.Duration(batchWaitMillis))
		}
	}
}

// 1 replica, 10ms batch
func BenchmarkModelUpdate_Models_10_Replicas_1_Batch_10ms(b *testing.B) {
	benchmarkModelUpdate(b, 10, 1, 1, 10)
}
func BenchmarkModelUpdate_Models_100_Replicas_1_Batch_10ms(b *testing.B) {
	benchmarkModelUpdate(b, 100, 1, 1, 10)
}
func BenchmarkModelUpdate_Models_1_000_Replicas_1_Batch_10ms(b *testing.B) {
	benchmarkModelUpdate(b, 1_000, 1, 1, 10)
}
func BenchmarkModelUpdate_Models_10_000_Replicas_1_Batch_10ms(b *testing.B) {
	benchmarkModelUpdate(b, 10_000, 1, 1, 10)
}

// 10 replicas, 10ms batch
func BenchmarkModelUpdate_Models_10_Replicas_10_Batch_10ms(b *testing.B) {
	benchmarkModelUpdate(b, 10, 1, 10, 10)
}
func BenchmarkModelUpdate_Models_100_Replicas_10_Batch_10ms(b *testing.B) {
	benchmarkModelUpdate(b, 100, 1, 10, 10)
}
func BenchmarkModelUpdate_Models_1_000_Replicas_10_Batch_10ms(b *testing.B) {
	benchmarkModelUpdate(b, 1_000, 1, 10, 10)
}
func BenchmarkModelUpdate_Models_10_000_Replicas_10_Batch_10ms(b *testing.B) {
	benchmarkModelUpdate(b, 10_000, 1, 10, 10)
}

// 1 replicas, 100ms batch
func BenchmarkModelUpdate_Models_10_Replicas_1_Batch_100ms(b *testing.B) {
	benchmarkModelUpdate(b, 1, 1, 1, 100)
}
func BenchmarkModelUpdate_Models_100_Replicas_1_Batch_100ms(b *testing.B) {
	benchmarkModelUpdate(b, 100, 1, 1, 100)
}
func BenchmarkModelUpdate_Models_1_000_Replicas_1_Batch_100ms(b *testing.B) {
	benchmarkModelUpdate(b, 1_000, 1, 1, 100)
}
func BenchmarkModelUpdate_Models_10_000_Replicas_1_Batch_100ms(b *testing.B) {
	benchmarkModelUpdate(b, 10_000, 1, 1, 100)
}

// 10 replicas, 100ms batch
func BenchmarkModelUpdate_Models_10_Replicas_10_Batch_100ms(b *testing.B) {
	benchmarkModelUpdate(b, 10, 1, 10, 100)
}
func BenchmarkModelUpdate_Models_100_Replicas_10_Batch_100ms(b *testing.B) {
	benchmarkModelUpdate(b, 100, 1, 10, 100)
}
func BenchmarkModelUpdate_Models_1_000_Replicas_10_Batch_100ms(b *testing.B) {
	benchmarkModelUpdate(b, 1_000, 1, 10, 100)
}
func BenchmarkModelUpdate_Models_10_000_Replicas_10_Batch_100ms(b *testing.B) {
	benchmarkModelUpdate(b, 10_000, 1, 10, 100)
}
