package filters

import (
	"testing"

	. "github.com/onsi/gomega"
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	"github.com/seldonio/seldon-core/scheduler/pkg/store"
)

func TestDeletedServerFilter(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name     string
		model    *store.ModelVersion
		server   *store.ServerSnapshot
		expected bool
	}
	serverName := "server1"
	model := store.NewModelVersion(
		&pb.Model{ModelSpec: &pb.ModelSpec{}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 1}},
		1,
		serverName,
		map[int]store.ReplicaStatus{3: {State: store.Loading}},
		false,
		store.ModelProgressing)
	tests := []test{
		{name: "DeletedServer", model: model, server: &store.ServerSnapshot{Name: serverName, Shared: true, ExpectedReplicas: 0}, expected: false},
		{name: "UnknownServerReplicas", model: model, server: &store.ServerSnapshot{Name: serverName, Shared: true, ExpectedReplicas: -1}, expected: true},
		{name: "ActiveServer", model: model, server: &store.ServerSnapshot{Name: serverName, Shared: true, ExpectedReplicas: 4}, expected: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filter := DeletedServerFilter{}
			ok := filter.Filter(test.model, test.server)
			g.Expect(ok).To(Equal(test.expected))
		})
	}
}
