/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package resources

import (
	"fmt"
	"os"

	"github.com/seldonio/seldon-core/operator/v2/pkg/cli"
)

type SeldonInferAPI struct {
	inferenceClient *cli.InferenceClient
	defaultCallOpts *cli.CallOptions
}

func NewSeldonInferAPI(host string) (*SeldonInferAPI, error) {
	ic, err := cli.NewInferenceClient(host, true)
	if err != nil {
		return nil, err
	}
	return &SeldonInferAPI{
		inferenceClient: ic,
		defaultCallOpts: &cli.CallOptions{
			InferProtocol: cli.InferRest,
			InferType:     cli.InferModel,
			StickySession: false,
			Iterations:    1,
		},
	}, nil
}

func (s *SeldonInferAPI) Infer(filename string, requestPath string) ([]byte, error) {
	callOpts := s.defaultCallOpts
	logOpts := &cli.LogOptions{}

	// Get infer type
	resourceMeta, err := getResource(filename)
	if err != nil {
		return nil, err
	}
	switch resourceMeta.gvk.Kind {
	case resourceModelKind:
		callOpts.InferType = cli.InferModel
	case resourcePipelineKind:
		callOpts.InferType = cli.InferPipeline
	default:
		return nil, fmt.Errorf("Cannot run infer for resource type %s in %s", resourceMeta.gvk.String(), filename)
	}

	// Get request
	request, err := os.ReadFile(requestPath)
	if err != nil {
		return nil, err
	}
	// Get infer protocol
	protocol, err := getInferRequestProtocol(request)
	if err != nil {
		return nil, err
	}
	callOpts.InferProtocol = protocol

	return removeIdFromResponse(s.inferenceClient.Infer(
		resourceMeta.name,
		request,
		[]string{},
		"",
		callOpts,
		logOpts,
	))
}
