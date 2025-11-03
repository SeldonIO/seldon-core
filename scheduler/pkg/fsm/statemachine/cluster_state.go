package statemachine

import pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

// ClusterState represents the state of things in the cluster needed for an event to be applied
type ClusterState struct {
	Model       map[string]*pb.ModelSnapshot
	Pipelines   map[string]*pb.PipelineSnapshot
	Experiments map[string]*pb.ExperimentSnapshot
}
