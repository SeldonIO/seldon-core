package resources

import (
	"fmt"
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

func (s *SeldonInferAPI) Infer(filename string, request string) error {
	callOpts := s.defaultCallOpts
	logOpts := &cli.LogOptions{}

	// Get infer type
	resourceMeta, err := getResource(filename)
	if err != nil {
		return err
	}
	switch resourceMeta.gvk.Kind {
	case resourceModelKind:
		callOpts.InferType = cli.InferModel
	case resourcePipelineKind:
		callOpts.InferType = cli.InferPipeline
	default:
		return fmt.Errorf("Unknown resource type in %s found %s", filename, resourceMeta.gvk.String())
	}

	// Get infer protocol
	protocol, err := getInferRequestProtocol(request)
	if err != nil {
		return err
	}
	callOpts.InferProtocol = protocol

	return s.inferenceClient.Infer(
		resourceMeta.name,
		[]byte(request),
		[]string{},
		"",
		callOpts,
		logOpts,
	)
}
