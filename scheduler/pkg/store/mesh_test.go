package store

import (
	"testing"
)

func TestReplicaStateToString(t *testing.T) {
	for _, state := range replicaStates {
		_ = state.String()
	}
}
