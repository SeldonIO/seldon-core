package store

import (
	"testing"

	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"

	. "github.com/onsi/gomega"
)

func TestReplicaStateToString(t *testing.T) {
	for _, state := range replicaStates {
		_ = state.String()
	}
}

func TestCleanCapabilities(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name     string
		in       []string
		expected []string
	}

	tests := []test{
		{
			name:     "misc",
			in:       []string{"mlserver", " foo ", " bar", "bar   "},
			expected: []string{"mlserver", "foo", "bar", "bar"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out := cleanCapabilities(test.in)
			g.Expect(out).To(Equal(test.expected))
		})
	}
}

func TestCreateSnapshot(t *testing.T) {
	g := NewGomegaWithT(t)

	server := &Server{
		name: "test",
		replicas: map[int]*ServerReplica{
			0: {
				inferenceSvc: "svc",
				loadedModels: map[ModelVersionID]bool{
					ModelVersionID{Name: "model1", Version: 1}: true,
					ModelVersionID{Name: "model2", Version: 2}: true,
				},
			},
		},
		kubernetesMeta: &pb.KubernetesMeta{Namespace: "default"},
	}

	snapshot := server.CreateSnapshot(false, true)

	server.replicas[1] = &ServerReplica{
		inferenceSvc: "svc",
		loadedModels: map[ModelVersionID]bool{
			ModelVersionID{Name: "model3", Version: 1}: true,
			ModelVersionID{Name: "model4", Version: 2}: true,
		},
	}
	server.name = "foo"
	server.kubernetesMeta.Namespace = "test"

	g.Expect(snapshot.Name).To(Equal("test"))
	g.Expect(len(snapshot.Replicas)).To(Equal(1))
	g.Expect(snapshot.KubernetesMeta.Namespace).To(Equal("default"))

}
