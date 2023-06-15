/*
Copyright 2023 Seldon Technologies Ltd.

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
