package filters

import (
	"testing"

	. "github.com/onsi/gomega"
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	"github.com/seldonio/seldon-core/scheduler/pkg/store"
)

func TestSharingFilter(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name     string
		model    *store.ModelVersion
		server   *store.ServerSnapshot
		expected bool
	}
	serverName := "server1"
	modelExplicitServer := store.NewModelVersion(
		&pb.ModelDetails{
			Name:   "model1",
			Server: &serverName,
		},
		serverName,
		map[int]store.ModelReplicaState{3: store.Loading},
		false,
		store.ModelProgressing)
	modelSharedServer := store.NewModelVersion(
		&pb.ModelDetails{
			Name:   "model1",
			Server: nil,
		},
		serverName,
		map[int]store.ModelReplicaState{3: store.Loading},
		false,
		store.ModelProgressing)
	tests := []test{
		{name: "ModelAndServerMatchNotShared", model: modelExplicitServer, server: &store.ServerSnapshot{Name: serverName, Shared: false}, expected: true},
		{name: "ModelAndServerMatchShared", model: modelExplicitServer, server: &store.ServerSnapshot{Name: serverName, Shared: true}, expected: true},
		{name: "ModelAndServerDontMatch", model: modelExplicitServer, server: &store.ServerSnapshot{Name: "foo", Shared: true}, expected: false},
		{name: "SharedModelAnyServer", model: modelSharedServer, server: &store.ServerSnapshot{Name: "foo", Shared: true}, expected: true},
		{name: "SharedModelNotSharedServer", model: modelSharedServer, server: &store.ServerSnapshot{Name: "foo", Shared: false}, expected: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filter := SharingServerFilter{}
			ok := filter.Filter(test.model, test.server)
			g.Expect(ok).To(Equal(test.expected))
		})
	}
}
