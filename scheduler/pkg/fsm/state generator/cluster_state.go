package state_generator

import pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

// ClusterState represents the state of things in the cluster needed for an event to be applied
type ClusterState struct {
	Models      map[string]*pb.ModelSnapshot
	Servers     map[string]*pb.ServerReplicaResources //todo: create a ServerSnapshot
	Pipelines   map[string]*pb.PipelineSnapshot
	Experiments map[string]*pb.ExperimentSnapshot
}
