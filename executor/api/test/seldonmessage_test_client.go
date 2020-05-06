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
)

func (s SeldonMessageTestClient) Status(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	return &payload.BytesPayload{Msg: []byte(TestClientStatusResponse)}, nil
}

func (s SeldonMessageTestClient) Metadata(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	return &payload.BytesPayload{Msg: []byte(TestClientMetadataResponse)}, nil
}

func (s SeldonMessageTestClient) Chain(ctx context.Context, modelName string, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	return msg, nil
}

func (s SeldonMessageTestClient) Unmarshall(msg []byte) (payload.SeldonPayload, error) {
	reqPayload := payload.BytesPayload{Msg: msg, ContentType: TestContentType}
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
	resp := &payload.ProtoPayload{Msg: msg.GetPayload().(*proto.Feedback).Request}
	return resp, nil
}
