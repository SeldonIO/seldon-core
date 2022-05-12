package hodometer

import (
	"fmt"
	"strings"
)

type MetricsLevel int

const (
	metricsLevelCluster MetricsLevel = iota
	metricsLevelResource
	metricsLevelFeature //nolint:varcheck
)

var supportedMetricsLevels = [...]string{
	"CLUSTER",
	"RESOURCE",
	"FEATURE",
}

func MetricsLevelFrom(level string) (MetricsLevel, error) {
	asUppercase := strings.ToUpper(level)
	for idx, sml := range supportedMetricsLevels {
		if sml == asUppercase {
			return MetricsLevel(idx), nil
		}
	}
	return -1, fmt.Errorf("level %s not recognised", level)
}

func (ml *MetricsLevel) String() string {
	if ml == nil {
		return "UNKNOWN"
	}
	return supportedMetricsLevels[int(*ml)]
}
