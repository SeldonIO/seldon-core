package pipeline

import (
	"fmt"
	"strings"

	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/resources"
)

func createResourceNameFromHeader(header string) (string, bool, error) {
	parts := strings.Split(header, ".")
	switch len(parts) {
	case 1:
		if len(parts[0]) > 0 {
			return header, true, nil
		}
	case 2:
		switch parts[1] {
		case resources.SeldonPipelineHeaderSuffix:
			return parts[0], false, nil
		case resources.SeldonModelHeaderSuffix:
			return parts[0], true, nil
		}
	}
	return "", false, fmt.Errorf(
		"Bad or missing header %s %s", resources.SeldonModelHeader, header)
}
