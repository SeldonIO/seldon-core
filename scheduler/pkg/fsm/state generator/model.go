package state_generator

import (
	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"google.golang.org/protobuf/proto"
)

// ========================================
// Model Snapshot Creation
// ========================================

// CreateModelSnapshot creates a new model snapshot with initial version
// Uses K8s generation as the starting version to ensure monotonic version numbers
func CreateModelSnapshot(model *pb.Model) *pb.ModelSnapshot {
	version := calculateInitialVersion(model)

	return &pb.ModelSnapshot{
		Versions: []*pb.ModelVersionStatus{
			createInitialModelVersion(model, version),
		},
		Deleted: false,
	}
}

// calculateInitialVersion extracts the K8s generation or defaults to 1
// This ensures version numbers never reset when recreating models
func calculateInitialVersion(model *pb.Model) uint32 {
	generation := model.GetMeta().GetKubernetesMeta().GetGeneration()
	return max(uint32(1), uint32(generation))
}

// createInitialModelVersion creates a fresh model version in unknown state
func createInitialModelVersion(model *pb.Model, version uint32) *pb.ModelVersionStatus {
	return &pb.ModelVersionStatus{
		Version:           version,
		ServerName:        "",
		KubernetesMeta:    model.GetMeta().GetKubernetesMeta(),
		ModelReplicaState: make(map[int32]*pb.ModelReplicaStatus),
		State: &pb.ModelStatus{
			State:               pb.ModelStatus_ModelStateUnknown,
			Reason:              "",
			AvailableReplicas:   0,
			UnavailableReplicas: 0,
			LastChangeTimestamp: nil,
			ModelGwState:        pb.ModelStatus_ModelCreate,
			ModelGwReason:       "",
		},
		ModelDefn: model,
	}
}

// ========================================
// Model Version Management
// ========================================

// GetLatestModelVersion returns the most recent version of a model
func GetLatestModelVersion(snapshot *pb.ModelSnapshot) *pb.ModelVersionStatus {
	if snapshot == nil || len(snapshot.Versions) == 0 {
		return nil
	}
	return snapshot.Versions[len(snapshot.Versions)-1]
}

// AddModelVersion appends a new version to a model snapshot
func AddModelVersion(snapshot *pb.ModelSnapshot, model *pb.Model, version uint32) *pb.ModelSnapshot {
	if snapshot == nil {
		return CreateModelSnapshot(model)
	}

	newVersion := createInitialModelVersion(model, version)
	snapshot.Versions = append(snapshot.Versions, newVersion)

	return snapshot
}

// calculateNextVersion determines the next version number based on the latest version
func calculateNextVersion(snapshot *pb.ModelSnapshot, model *pb.Model) uint32 {
	latestVersion := GetLatestModelVersion(snapshot)
	if latestVersion == nil {
		return calculateInitialVersion(model)
	}

	// Use the greater of: latest version + 1, or K8s generation
	generation := model.GetMeta().GetKubernetesMeta().GetGeneration()
	nextVersion := latestVersion.Version + 1

	return max(nextVersion, uint32(generation))
}

// ========================================
// State Checking Utilities
// ========================================

// IsModelFullyInactive checks if ALL versions of a model are inactive
func IsModelFullyInactive(snapshot *pb.ModelSnapshot) bool {
	if snapshot == nil || len(snapshot.Versions) == 0 {
		return true
	}

	for _, version := range snapshot.Versions {
		if !IsModelVersionInactive(version) {
			return false
		}
	}

	return true
}

// IsModelVersionInactive checks if a model version has no active replicas
func IsModelVersionInactive(version *pb.ModelVersionStatus) bool {
	if version == nil || len(version.ModelReplicaState) == 0 {
		return true
	}

	for _, replicaStatus := range version.ModelReplicaState {
		if !IsReplicaInactive(replicaStatus.State) {
			return false
		}
	}

	return true
}

// IsReplicaInactive checks if a replica is in an inactive state
func IsReplicaInactive(state pb.ModelReplicaStatus_ModelReplicaState) bool {
	switch state {
	case pb.ModelReplicaStatus_Unloaded,
		pb.ModelReplicaStatus_UnloadFailed,
		pb.ModelReplicaStatus_ModelReplicaStateUnknown,
		pb.ModelReplicaStatus_LoadFailed:
		return true
	default:
		return false
	}
}

// IsReplicaActive checks if a replica is in an active/loading state
func IsReplicaActive(state pb.ModelReplicaStatus_ModelReplicaState) bool {
	return !IsReplicaInactive(state)
}

// ========================================
// Model Snapshot Generation
// ========================================

// generateModelSnapshot creates or updates a model snapshot based on the request
// Handles:
// - New models: creates fresh snapshot
// - Deleted models: adds new version if all replicas are inactive
// - Existing models: updates with new definition
// - checks if deployment spec differs on model and creates a
func (cs *ClusterState) generateModelSnapshot(requestedModel *pb.Model) *pb.ModelSnapshot {
	modelName := requestedModel.GetMeta().GetName()

	// Check if model exists
	currentSnapshot, exists := cs.Model[modelName]
	if !exists {
		// Brand new model - create initial snapshot
		return CreateModelSnapshot(requestedModel)
	}

	// Model exists - check if it was previously deleted
	if currentSnapshot.GetDeleted() {
		return cs.handleDeletedModelRecreation(currentSnapshot, requestedModel)
	}

	// Case 3: Existing active model - check what changed
	latestVersion := GetLatestModelVersion(currentSnapshot)
	if latestVersion == nil {
		// Shouldn't happen, but handle gracefully
		return CreateModelSnapshot(requestedModel)
	}

	// Compare current model definition with requested model
	currentModel := latestVersion.ModelDefn
	equality := store.ModelEqualityCheck(currentModel, requestedModel)

	// Case 3a: No changes - return clone of current state
	if equality.Equal {
		return proto.Clone(currentSnapshot).(*pb.ModelSnapshot)
	}

	// Clone for modifications
	outputSnap := proto.Clone(currentSnapshot).(*pb.ModelSnapshot)
	latestInOutput := GetLatestModelVersion(outputSnap)
	if latestInOutput == nil {
		// Safety check - shouldn't happen since we checked above
		return CreateModelSnapshot(requestedModel)
	}

	// Case 3b: DeploymentSpec changed - requires new version (rolling update)
	// This triggers replica changes, so we need a new version to track separately
	if equality.DeploymentSpecDiffers {
		nextVersion := calculateNextVersion(currentSnapshot, requestedModel)
		outputSnap = AddModelVersion(outputSnap, requestedModel, nextVersion)
		return outputSnap
	}

	// Case 3c: ModelSpec changed - update in place
	// This is configuration changes that don't require new replicas
	if equality.ModelSpecDiffers {
		latestInOutput.ModelDefn = requestedModel
		// Note: Keep existing replicas, just update config
	}

	// Case 3d: Only metadata changed - update in place
	if equality.MetaDiffers {
		latestInOutput.KubernetesMeta = requestedModel.GetMeta().GetKubernetesMeta()
		latestInOutput.ModelDefn = requestedModel
	}

	return outputSnap
}

// handleDeletedModelRecreation handles recreating a model that was previously deleted
func (cs *ClusterState) handleDeletedModelRecreation(
	snapshot *pb.ModelSnapshot,
	requestedModel *pb.Model,
) *pb.ModelSnapshot {
	// Only allow recreation if all versions are fully inactive
	if !IsModelFullyInactive(snapshot) {
		// Still has active replicas - cannot recreate yet
		// Return existing snapshot unchanged

		//todo: this used to be an error
		return snapshot
	}

	// All inactive - safe to add a new version
	nextVersion := calculateNextVersion(snapshot, requestedModel)
	updatedSnapshot := AddModelVersion(snapshot, requestedModel, nextVersion)
	updatedSnapshot.Deleted = false // Mark as no longer deleted

	return updatedSnapshot
}

// updateExistingModel updates an existing active model with new definition
func (cs *ClusterState) updateExistingModel(
	snapshot *pb.ModelSnapshot,
	requestedModel *pb.Model,
) *pb.ModelSnapshot {
	latestVersion := GetLatestModelVersion(snapshot)
	if latestVersion == nil {
		// Shouldn't happen, but handle gracefully
		return CreateModelSnapshot(requestedModel)
	}

	// Check if this is actually a new version (K8s generation changed)
	currentGeneration := latestVersion.GetKubernetesMeta().GetGeneration()
	requestedGeneration := requestedModel.GetMeta().GetKubernetesMeta().GetGeneration()

	if requestedGeneration > currentGeneration {
		// New K8s generation - create new version
		nextVersion := calculateNextVersion(snapshot, requestedModel)
		return AddModelVersion(snapshot, requestedModel, nextVersion)
	}

	// Same generation - update model definition in place
	latestVersion.ModelDefn = requestedModel
	latestVersion.KubernetesMeta = requestedModel.GetMeta().GetKubernetesMeta()

	return snapshot
}
