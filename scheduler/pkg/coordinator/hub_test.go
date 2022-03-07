package coordinator

import (
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestNewEventHub(t *testing.T) {
	type test struct {
		name           string
		expectedTopics []string
	}

	tests := []test{
		{
			name:           "Should register two topics",
			expectedTopics: []string{topicModelEvents, topicExperimentEvents},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := log.New()
			h, err := NewEventHub(l)
			require.Nil(t, err)
			require.ElementsMatch(t, tt.expectedTopics, h.bus.Topics())
		})
	}
}
