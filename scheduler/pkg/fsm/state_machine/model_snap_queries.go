package state_machine

// Here we will have Model query methods (GetStatus, IsReady, etc)

// IsModelFullyInactive checks if ALL versions of a model are inactive
func (ms *ModelSnapshot) IsModelFullyInactive() bool {
	if ms == nil || len(ms.Versions) == 0 {
		return true
	}

	for _, version := range ms.Versions {
		if NewModelVersion(version).Active() == false {
			return true
		}
	}

	return false
}

func (ms *ModelSnapshot) GetLatestModelVersionStatus() *ModelVersionStatus {
	if ms == nil || len(ms.Versions) == 0 {
		return nil
	}

	return NewModelVersion(ms.Versions[len(ms.Versions)-1])
}
