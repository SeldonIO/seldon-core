package state_machine

import (
	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"google.golang.org/protobuf/proto"
)

// Here we will have the Model snapshot wrapper and basic methods

type ModelSnapshot struct {
	*pb.ModelSnapshot
}

func NewModelSnapshot(proto *pb.ModelSnapshot) *ModelSnapshot {
	return &ModelSnapshot{ModelSnapshot: proto}
}

// CreateModelSnapshot creates a new model snapshot with initial version
// Uses K8s generation as the starting version to ensure monotonic version numbers
func CreateModelSnapshot(model *pb.Model) *ModelSnapshot {
	version := calculateInitialVersion(model)

	return NewModelSnapshot(&pb.ModelSnapshot{
		Versions: []*pb.ModelVersionStatus{
			createInitialModelVersion(model, version).ModelVersionStatus,
		},
		Deleted: false,
	})
}

func (ms *ModelSnapshot) NewVersion(model *pb.Model) *ModelSnapshot {
	lastModelVersion := ms.GetLatestModelVersionStatus()
	if lastModelVersion == nil {
		return ms
	}

	ms.Versions = append(ms.Versions, createInitialModelVersion(model, lastModelVersion.Version+1).ModelVersionStatus)

	return ms
}

// calculateInitialVersion extracts the K8s generation or defaults to 1
// This ensures version numbers never reset when recreating models
func calculateInitialVersion(model *pb.Model) uint32 {
	generation := model.GetMeta().GetKubernetesMeta().GetGeneration()
	return max(uint32(1), uint32(generation))
}

// ========================================
// Model Snapshot Generation
// ========================================

// todo: this might have to a function instead of a method of cluster state or even part of a struct for model operations
// generateModelSnapshot creates or updates a model snapshot based on the request
// Handles:
// - New models: creates fresh snapshot
// - Deleted models: adds new version if all replicas are inactive
// - Existing models: updates with new definition
// - checks if deployment spec differs on model and creates a new model version
func (cs *ClusterState) generateModelSnapshot(requestedModel *pb.Model) *ModelSnapshot {
	modelName := requestedModel.GetMeta().GetName()

	// Check if model exists
	currentModelSnap, exists := cs.Models[modelName]
	if !exists {
		// Brand new model - create initial snapshot
		return CreateModelSnapshot(requestedModel)
	}

	// Model exists - check if it was previously deleted
	if currentModelSnap.GetDeleted() {
		return cs.handleDeletedModelRecreation(currentModelSnap, requestedModel)
	}

	// Case 3: Existing active model - check what changed
	currentLatestModelVer := currentModelSnap.GetLatestModelVersionStatus()
	if currentLatestModelVer == nil {
		// Shouldn't happen, but handle gracefully
		return CreateModelSnapshot(requestedModel)
	}

	// Compare current latest model ver definition with requested model
	currentModel := currentLatestModelVer.ModelDefn
	equality := store.ModelEqualityCheck(currentModel, requestedModel)

	// Case 3a: No changes - return clone of current state
	if equality.Equal {
		return currentModelSnap
	}

	// Clone for modifications
	outputSnap := NewModelSnapshot(proto.Clone(currentModelSnap.ModelSnapshot).(*pb.ModelSnapshot))
	latestInOutput := outputSnap.GetLatestModelVersionStatus()
	if latestInOutput == nil {
		// Safety check - shouldn't happen since we checked above
		return CreateModelSnapshot(requestedModel)
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

// handleDeletedModelRecreation handles recreating a model that was previously deleted
func (cs *ClusterState) handleDeletedModelRecreation(
	snapshot *ModelSnapshot,
	requestedModel *pb.Model,
) *ModelSnapshot {
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
