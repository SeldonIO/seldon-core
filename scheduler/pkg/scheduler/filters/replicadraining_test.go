package filters

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/scheduler/pkg/store"
)

func TestReplicaDrainingFilter(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name       string
		isDraining bool
		expected   bool
	}

	tests := []test{
		{name: "WithDraining", isDraining: true, expected: false},
		{name: "NoDraining", isDraining: false, expected: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filter := ReplicaDrainingFilter{}
			replica := store.ServerReplica{}
			if test.isDraining {
				replica.SetIsDraining()
			}
			ok := filter.Filter(nil, &replica)
			g.Expect(ok).To(Equal(test.expected))
		})
	}
}
