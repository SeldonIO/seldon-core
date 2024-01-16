/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

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
