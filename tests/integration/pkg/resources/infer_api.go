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
	"fmt"
	"github.com/seldonio/seldon-core/operator/v2/pkg/cli"
	"os"
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
