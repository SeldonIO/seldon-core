package resources

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type SeldonBackendAPI interface {
	Load(filename string) error
	UnLoad(filename string) error
	IsLoaded(filename string) (bool, error)
	Infer(filename string, request string) error
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
