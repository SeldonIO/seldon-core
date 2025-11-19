/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package state_machine

import (
	"fmt"

	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/events"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine/model"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine/server"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"google.golang.org/protobuf/proto"
)

// Model represents the state machine for Model
type Model interface {
	ApplyLoadModel(current ClusterState, event events.LoadModel) (ClusterState, error)
	ApplyUnloadModel(current ClusterState, request *pb.UnloadModelRequest) (ClusterState, error)
	ApplyModelStateLoadRequested(current ClusterState) (ClusterState, error)
}

// ApplyLoadModel applies business logic for loading a model
// Pure function: same inputs → same outputs
func (sm *StateMachine) ApplyLoadModel(
	current ClusterState,
	event events.LoadModel,
) (ClusterState, error) {
	modelName := event.GetModel().GetMeta().GetName()

	if modelName == "" {
		return current, fmt.Errorf("model name is empty")
	}

	// calculate the future model state
	futureModelState := getModelSnapLoadModel(current, event.Model)

	// calculate server state
	futureServerState := getServerSnapLoadModel(current, event.Model)

	// calculate the future pipelines states
	// todo: if the model existed but its definition changed we need to update the status of a pipeline
	// todo: properly check the pipeline logic for this

	// calculate the future experiment states
	//todo: this might follow in line with the changes of pipelines

	// todo: think more about this but I think this is better than cloning a very big struct
	futureCluster := ClusterState{
		Models: map[string]*model.Snapshot{
			modelName: futureModelState,
		},
		Servers: map[string]*server.Snapshot{
			futureServerState.Name: futureServerState,
		},
	}

	return futureCluster, nil
}

func (sm *StateMachine) ApplyUnloadModel(current ClusterState, request *pb.UnloadModelRequest) (ClusterState, error) {
	//todo:

	// similarly to the load request this should mark all the models or the latest one as unloaded?

	// this will also change the status of pipelines and experiments that use the model

	panic("implement me")
	return ClusterState{}, nil
}

func (sm *StateMachine) ApplyModelStateLoadRequested(current ClusterState) (ClusterState, error) {
	//todo:

	// this will be a state transition of the model this event can also affect the status of pipelines and experiments
	// that use the model
	panic("implement me")
	return ClusterState{}, nil
}

// ========================================
// Model Snapshot Generation
// ========================================

// todo: this might have to a function instead of a method of cluster state or even part of a struct for model operations
// todo: move to handler
// getModelSnapLoadModel creates or updates a model snapshot based on the request
// Handles:
// - New models: creates fresh snapshot
// - Deleted models: adds new version if all replicas are inactive
// - Existing models: updates with new definition
// - checks if deployment spec differs on model and creates a new model version
func getModelSnapLoadModel(currentState ClusterState, requestedModel *pb.Model) *model.Snapshot {
	modelName := requestedModel.GetMeta().GetName()

	// Check if model exists
	currentModelSnap, exists := currentState.Models[modelName]
	if !exists {
		// Brand new model - create initial snapshot
		return model.CreateSnapshotFromModel(requestedModel)
	}

	// Model exists - check if it was previously deleted
	if currentModelSnap.GetDeleted() {
		return handleDeletedModelRecreation(currentModelSnap, requestedModel)
	}

	// Case 3: Existing active model - check what changed
	currentLatestModelVer := currentModelSnap.GetLatestModelVersionStatus()
	if currentLatestModelVer == nil {
		// Shouldn't happen, but handle gracefully
		return model.CreateSnapshotFromModel(requestedModel)
	}

	//todo: I think I need to check model status

	// Compare current latest model ver definition with requested model
	currentModel := currentLatestModelVer.ModelDefn
	equality := store.ModelEqualityCheck(currentModel, requestedModel)

	// Case 3a: No changes - return clone of current state
	if equality.Equal {
		return currentModelSnap
	}

	// Clone for modifications
	outputSnap := model.NewSnapshot(proto.Clone(currentModelSnap.ModelSnapshot).(*pb.ModelSnapshot))
	latestInOutput := outputSnap.GetLatestModelVersionStatus()
	if latestInOutput == nil {
		// Safety check - shouldn't happen since we checked above
		return model.CreateSnapshotFromModel(requestedModel)
	}

	// Case 3b: DeploymentSpec changed - requires new version (rolling update)
	// This triggers replica changes, so we need a new version to track separately
	if equality.DeploymentSpecDiffers {
		return currentModelSnap.NewVersion(requestedModel)
	}

	// Case 3c: ModelSpec changed - update in place
	// This is configuration changes that don't require new replicas
	if equality.ModelSpecDiffers {
		outputSnap.GetLatestModelVersionStatus().ModelDefn = requestedModel
		// Note: Keep existing replicas, just update config
	}

	// Case 3d: Only metadata changed - update in place
	if equality.MetaDiffers {
		outputSnap.GetLatestModelVersionStatus().KubernetesMeta = requestedModel.GetMeta().GetKubernetesMeta()
		outputSnap.GetLatestModelVersionStatus().ModelDefn = requestedModel
	}

	return outputSnap
}

// getServerSnapLoadModel calculates the server snapshot for a load model request
func getServerSnapLoadModel(currentState ClusterState, requestedModel *pb.Model) *server.Snapshot {
	// things we need to do

	/*
		 - filter servers
		    - if no filter return emtpy server snapshot (meaning that we could not schedule the model
			- could even mark the model as failed deployment
		- get model desired replicas and min replicas
		- sort servers (i guess this is by availability of who has the highest)
		- on a rolling update this will need a few trips (probably the loadagent request when the model is confirmed
		  as loaded triggers the next replica to be loaded
		- reserve memory for each replica but with status memory reserved and only load request for the subsequent deploys

	*/

	return nil
}

// handleDeletedModelRecreation handles recreating a model that was previously deleted
func handleDeletedModelRecreation(
	snapshot *model.Snapshot,
	requestedModel *pb.Model,
) *model.Snapshot {
	// Only allow recreation if all versions are fully inactive

	latestModelVersion := snapshot.GetLatestModelVersionStatus()
	if latestModelVersion == nil {
		return snapshot
	}

	if latestModelVersion.Active() {
		// is being mark as deleted and still has pending  state to process do nothing and return snap
		return snapshot
	}

	// All inactive - safe to add a new version
	futureSnap := snapshot.NewVersion(requestedModel)
	futureSnap.Deleted = false // Mark as no longer deleted

	return futureSnap
}
