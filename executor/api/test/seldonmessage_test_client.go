package test

import (
	"context"
	// "errors"
	"fmt"
	"io"

	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"github.com/seldonio/seldon-core/executor/api/payload"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
)

type SeldonMessageTestClient struct {
	ChosenRoute      int
	MetadataResponse payload.SeldonPayload
	ModelMetadataMap map[string]payload.ModelMetadata
	ErrMethod        *v1.PredictiveUnitMethod
	Err              error
	ErrPayload       payload.SeldonPayload
}

const (
	TestClientStatusResponse = `{"status":"ok"}`
	TestContentType          = "application/json"
	TestGraphMeta            = `{
		"name": "predictor-name",
		"models": {
			"model-1": {
				"name": "model-1",
				"platform": "platform-name",
				"versions": ["model-version"],
				"inputs": [{"datatype": "BYTES", "name": "input", "shape": [1, 5]}],
				"outputs": [{"datatype": "BYTES", "name": "output", "shape": [1, 3]}]
			},
			"model-2": {
				"name": "model-2",
				"platform": "platform-name",
				"versions": ["model-version"],
				"inputs": [{"datatype": "BYTES", "name": "input", "shape": [1, 3]}],
				"outputs": [{"datatype": "BYTES", "name": "output", "shape": [3]}]}
			},
		"graphinputs": [{"datatype": "BYTES", "name": "input", "shape": [1, 5]}],
		"graphoutputs": [{"datatype": "BYTES", "name": "output", "shape": [3]}]
	}`
)

func (s SeldonMessageTestClient) IsGrpc() bool {
	return true
}

func (s SeldonMessageTestClient) Status(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	return &payload.BytesPayload{Msg: []byte(TestClientStatusResponse)}, nil
}

func (s SeldonMessageTestClient) Metadata(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	if s.MetadataResponse == nil {
		return nil, fmt.Errorf("Metadata %s not present in test client", modelName)
	}
	return s.MetadataResponse, nil
}

func (s SeldonMessageTestClient) ModelMetadata(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.ModelMetadata, error) {
	output, present := s.ModelMetadataMap[modelName]
	if !present {
		return payload.ModelMetadata{}, fmt.Errorf("Metadata %s not present in test client", modelName)
	}
	return output, nil
}

func (s SeldonMessageTestClient) Chain(ctx context.Context, modelName string, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	return msg, nil
}

func (s SeldonMessageTestClient) Unmarshall(msg []byte, contentType string) (payload.SeldonPayload, error) {
	reqPayload := payload.BytesPayload{Msg: msg, ContentType: contentType}
	return &reqPayload, nil
}

func (s SeldonMessageTestClient) Marshall(out io.Writer, msg payload.SeldonPayload) error {
	_, err := out.Write(msg.GetPayload().([]byte))
	return err
}

func (s SeldonMessageTestClient) CreateErrorPayload(err error) payload.SeldonPayload {
	respFailed := payload.BytesPayload{Msg: []byte(err.Error()), ContentType: TestContentType}
	return &respFailed
}

func (s SeldonMessageTestClient) Predict(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	if s.ErrMethod != nil && *s.ErrMethod == v1.TRANSFORM_INPUT {
		return s.ErrPayload, s.Err
	}
	return msg, nil
}

func (s SeldonMessageTestClient) TransformInput(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	if s.ErrMethod != nil && *s.ErrMethod == v1.TRANSFORM_INPUT {
		return s.ErrPayload, s.Err
	}
	return msg, nil
}

func (s SeldonMessageTestClient) Route(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (int, error) {
	return s.ChosenRoute, nil
}

func (s SeldonMessageTestClient) Combine(ctx context.Context, modelName string, host string, port int32, msgs []payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	return msgs[0], nil
}

func (s SeldonMessageTestClient) TransformOutput(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	return msg, nil
}

func (s SeldonMessageTestClient) Feedback(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	if s.ErrMethod != nil && *s.ErrMethod == v1.SEND_FEEDBACK {
		return nil, s.Err
	}
	protoFeedback, ok := msg.GetPayload().(*proto.Feedback)
	if ok {
		resp := &payload.ProtoPayload{Msg: protoFeedback.Request}
		return resp, nil
	}
	return msg, nil
}
