/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

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
			expectedTopics: []string{topicModelEvents, topicExperimentEvents, topicPipelineEvents},
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
