package state_machine

import pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

// generatePipelineSnapshot will generate the next state of pipelines when a model state is change
func (cs *ClusterState) generatePipelineSnapshotFromModel(requestedModel *pb.ModelSnapshot) []*pb.PipelineSnapshot {
	// check if model deleted

	requestedModel.Deleted
}
