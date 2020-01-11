package test

import (
	"context"
	"github.com/seldonio/seldon-core/executor/api/client"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/operator/apis/machinelearning/v1"
	"io"
	"net/http"
	"testing"
)

type SeldonMessageTestClient struct {
	t           *testing.T
	chosenRoute int
	errMethod   *v1.PredictiveUnitMethod
	err         error
}

const (
	TestClientStatusResponse   = `{"status":"ok"}`
	TestClientMetadataResponse = `{"metadata":{"name":"mymodel"}}`
)

func (s SeldonMessageTestClient) Status(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	return &payload.BytesPayload{Msg: []byte(TestClientStatusResponse)}, nil
}

func (s SeldonMessageTestClient) Metadata(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	return &payload.BytesPayload{Msg: []byte(TestClientMetadataResponse)}, nil
}

func (s SeldonMessageTestClient) Chain(ctx context.Context, modelName string, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	return msg, nil
}

func (s SeldonMessageTestClient) Unmarshall(msg []byte) (payload.SeldonPayload, error) {
	reqPayload := payload.BytesPayload{Msg: msg, ContentType: "application/json"}
	return &reqPayload, nil
}

func (s SeldonMessageTestClient) Marshall(out io.Writer, msg payload.SeldonPayload) error {
	_, err := out.Write(msg.GetPayload().([]byte))
	return err
}

func (s SeldonMessageTestClient) CreateErrorPayload(err error) payload.SeldonPayload {
	respFailed := proto.SeldonMessage{Status: &proto.Status{Code: http.StatusInternalServerError, Info: err.Error()}}
	res := payload.ProtoPayload{Msg: &respFailed}
	return &res
}

func (s SeldonMessageTestClient) Predict(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	s.t.Logf("Predict %s %d", host, port)
	if s.errMethod != nil && *s.errMethod == v1.TRANSFORM_INPUT {
		return nil, s.err
	}
	return msg, nil
}

func (s SeldonMessageTestClient) TransformInput(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	s.t.Logf("TransformInput %s %d", host, port)
	if s.errMethod != nil && *s.errMethod == v1.TRANSFORM_INPUT {
		return nil, s.err
	}
	return msg, nil
}

func (s SeldonMessageTestClient) Route(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload) (int, error) {
	s.t.Logf("Route %s %d", host, port)
	return s.chosenRoute, nil
}

func (s SeldonMessageTestClient) Combine(ctx context.Context, modelName string, host string, port int32, msgs []payload.SeldonPayload) (payload.SeldonPayload, error) {
	s.t.Logf("Combine %s %d", host, port)
	return msgs[0], nil
}

func (s SeldonMessageTestClient) TransformOutput(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	s.t.Logf("TransformOutput %s %d", host, port)
	return msg, nil
}

func (s SeldonMessageTestClient) Feedback(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	s.t.Logf("Feedback %s %d", host, port)
	if s.errMethod != nil && *s.errMethod == v1.SEND_FEEDBACK {
		return nil, s.err
	}
	resp := &payload.ProtoPayload{Msg: msg.GetPayload().(*proto.Feedback).Request}
	return resp, nil
}

func NewSeldonMessageTestClient(t *testing.T, chosenRoute int, errMethod *v1.PredictiveUnitMethod, err error) client.SeldonApiClient {
	client := SeldonMessageTestClient{
		t:           t,
		chosenRoute: chosenRoute,
		errMethod:   errMethod,
		err:         err,
	}
	return &client
}
