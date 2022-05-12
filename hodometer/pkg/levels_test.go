package hodometer

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func levelPtr(ml MetricsLevel) *MetricsLevel {
	return &ml
}

func TestString(t *testing.T) {
	type test struct {
		name     string
		level    *MetricsLevel
		expected string
	}

	tests := []test{
		{
			name:     "cluster level",
			level:    levelPtr(metricsLevelCluster),
			expected: "CLUSTER",
		},
		{
			name:     "resource level",
			level:    levelPtr(metricsLevelResource),
			expected: "RESOURCE",
		},
		{
			name:     "feature level",
			level:    levelPtr(metricsLevelFeature),
			expected: "FEATURE",
		},
		{
			name:     "nil level",
			level:    nil,
			expected: "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.level.String()
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestMetricsLevelFrom(t *testing.T) {
	type test struct {
		name          string
		levelName     string
		expected      MetricsLevel
		expectedError error
	}

	tests := []test{
		{
			name:          "cluster level",
			levelName:     "CLUSTER",
			expected:      metricsLevelCluster,
			expectedError: nil,
		},
		{
			name:          "resource level",
			levelName:     "RESOURCE",
			expected:      metricsLevelResource,
			expectedError: nil,
		},
		{
			name:          "feature level",
			levelName:     "FEATURE",
			expected:      metricsLevelFeature,
			expectedError: nil,
		},
		{
			name:          "not a metrics level",
			levelName:     "asdf",
			expected:      -1,
			expectedError: errors.New("level asdf not recognised"),
		},
		{
			name:          "lowercase level name",
			levelName:     "feature",
			expected:      metricsLevelFeature,
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level, err := MetricsLevelFrom(tt.levelName)
			if tt.expectedError != nil {
				require.Error(t, err)
			} else {
				require.Nil(t, err)
			}
			require.Equal(t, tt.expected, level)
		})
	}
}
