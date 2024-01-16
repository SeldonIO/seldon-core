/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package resources

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type SeldonBackendAPI interface {
	Load(filename string) error
	Unload(filename string) error
	IsLoaded(filename string) (bool, error)
	Infer(filename string, request string) ([]byte, error)
}

func checkSeldonRunningLocally() (bool, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return false, err
	}
	// get all running containers
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return false, err
	}
	for _, container := range containers {
		for _, name := range container.Names {
			if name == "/scv2-scheduler-1" {
				return true, nil
			}
		}
	}
	return false, nil
}

func NewSeldonBackendAPI() (SeldonBackendAPI, error) {
	seldonRunningInDocker, err := checkSeldonRunningLocally()
	if err != nil {
		return nil, err
	}
	if seldonRunningInDocker {
		return NewSeldonAPI()
	} else {
		return NewSeldonK8sAPI()
	}
}
