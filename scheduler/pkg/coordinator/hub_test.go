/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
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
