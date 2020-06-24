package test

import (
	"context"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"io"
)

type SeldonMessageTestClient struct {
	ChosenRoute int
	ErrMethod   *v1.PredictiveUnitMethod
	Err         error
	ErrPayload  payload.SeldonPayload
}

const (
	TestClientStatusResponse   = `{"status":"ok"}`
	TestClientMetadataResponse = `{"metadata":{"name":"mymodel"}}`
	TestContentType            = "application/json"
	TestGraphMeta              = `{
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

var metadataMap = map[string]string{
	"mymodel": TestClientMetadataResponse,
	"model-1": `{
		"name": "model-1",
		"versions": ["model-version"],
		"platform": "platform-name",
		"inputs": [{"name": "input", "datatype": "BYTES", "shape": [1, 5]}],
		"outputs": [{"name": "output", "datatype": "BYTES", "shape": [1, 3]}]
    }`,
	"model-2": `{
		"name": "model-2",
		"versions": ["model-version"],
		"platform": "platform-name",
		"inputs": [{"name": "input", "datatype": "BYTES", "shape": [1, 3]}],
		"outputs": [{"name": "output", "datatype": "BYTES", "shape": [3]}]
    }`,
	"model-a1": `{
        "name": "model-a1",
        "versions": ["model-version"],
        "platform": "platform-name",
        "inputs": [{"name": "input", "datatype": "BYTES", "shape": [1, 5]}],
        "outputs": [{"name": "output", "datatype": "BYTES", "shape": [1, 10]}]
    }`,
	"model-a2": `{
        "name": "model-a2",
        "versions": ["model-version"],
        "platform": "platform-name",
        "inputs": [{"name": "input", "datatype": "BYTES", "shape": [1, 5]}],
        "outputs": [{"name": "output", "datatype": "BYTES", "shape": [1, 20]}]
    }`,
	"model-b1": `{
        "name": "model-b1",
        "versions": ["model-version"],
        "platform": "platform-name",
        "inputs": [{"name": "input", "datatype": "BYTES", "shape": [1, 5]}],
        "outputs": [{"name": "output", "datatype": "BYTES", "shape": [1, 10]}]
    }`,
	"model-router": `{
        "name": "model-router",
        "versions": ["model-version"],
        "platform": "platform-name"
    }`,
	"model-combiner": `{
        "name": "model-combiner",
        "versions": ["model-version"],
        "platform": "platform-name",
        "inputs": [
            {"name": "input-1", "datatype": "BYTES", "shape": [1, 10]},
            {"name": "input-2", "datatype": "BYTES", "shape": [1, 20]}
        ],
        "outputs": [{"name": "combined output", "datatype": "BYTES", "shape": [3]}]
    }`,
	"model-v1-array": `{
		"name": "model-v1-array",
		"versions": ["model-version"],
		"platform": "platform-name",
		"inputs": {"datatype": "array", "shape": [2, 2]},
		"outputs": {"datatype": "array", "shape": [1]}
	}`,
	"model-v1-jsondata": `{
		"name": "model-v1-jsondata",
		"versions": ["model-version"],
		"platform": "platform-name",
		"inputs": {"datatype": "jsonData"},
		"outputs": {"datatype": "jsonData", "schema": {"custom": "definition"}}
	}`,
	"model-v1-array-string-mix": `{
		"name": "model-v1-array-string-mix",
		"versions": ["model-version"],
		"platform": "platform-name",
		"inputs": {"datatype": "array", "shape": [2, 2]},
		"outputs": {"datatype": "strData"}
	}`,
}

func (s SeldonMessageTestClient) Status(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	return &payload.BytesPayload{Msg: []byte(TestClientStatusResponse)}, nil
}

func (s SeldonMessageTestClient) Metadata(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	return &payload.BytesPayload{ContentType: "application/json", Msg: []byte(metadataMap[modelName])}, nil
}

func (s SeldonMessageTestClient) ModelMetadata(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.ModelMetadata, error) {
	return payload.ModelMetadata{}, nil
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
	if ok == true {
		resp := &payload.ProtoPayload{Msg: protoFeedback.Request}
		return resp, nil
	}
	return msg, nil
}
