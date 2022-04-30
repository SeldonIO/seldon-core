package store

import (
	"testing"

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
