package cli

import (
	"fmt"

	"github.com/seldonio/seldon-core/operator/v2/pkg/cli"
)

func getInferProtocol(inferMode string) (cli.InferProtocol, error) {
	switch inferMode {
	case "rest":
		return cli.InferRest, nil
	case "grpc":
		return cli.InferGrpc, nil
	default:
		return cli.InferUnknown, fmt.Errorf("Unknown infer mode %s", inferMode)
	}
}
