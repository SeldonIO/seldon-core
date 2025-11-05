package state_machine

// IsInactive checks if a model version status has no active replicas
func (mvs *ModelVersionStatus) Active() bool {
	if mvs == nil || len(mvs.ModelReplicaState) == 0 {
		return false
	}

	for _, state := range mvs.ModelReplicaState {
		if NewModelReplicaStatus(state).Active() {
			return true
		}
	}
	return false
}

func (mvs *ModelVersionStatus) HasServer() bool {
	return mvs.ServerName != ""
}
