package test

import (
	"github.com/golang/protobuf/jsonpb"
	"github.com/seldonio/seldon-core/executor/api/client"
	"github.com/seldonio/seldon-core/executor/api/grpc/proto"
	"github.com/seldonio/seldon-core/executor/api/machinelearning/v1alpha2"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"io"
	"net/http"
	"testing"
)

type SeldonMessageTestClient struct {
	t           *testing.T
	chosenRoute int
	errMethod   *v1alpha2.PredictiveUnitMethod
	err         error
}

func (s SeldonMessageTestClient) Unmarshall(msg []byte) (payload.SeldonPayload, error) {
	var sm proto.SeldonMessage
	value := string(msg)
	if err := jsonpb.UnmarshalString(value, &sm); err != nil {
		return nil, err
	}
	reqPayload := payload.SeldonMessagePayload{Msg: &sm}
	return &reqPayload, nil
}

func (s SeldonMessageTestClient) Marshall(out io.Writer, msg payload.SeldonPayload) error {
	ma := jsonpb.Marshaler{}
	return ma.Marshal(out, msg.GetPayload().(*proto.SeldonMessage))
}

func (s SeldonMessageTestClient) CreateErrorPayload(err error) payload.SeldonPayload {
	respFailed := proto.SeldonMessage{Status: &proto.Status{Code: http.StatusInternalServerError, Info: err.Error()}}
	res := payload.SeldonMessagePayload{Msg: &respFailed}
	return &res
}

func (s SeldonMessageTestClient) Predict(host string, port int32, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	s.t.Logf("Predict %s %d", host, port)
	if s.errMethod != nil && *s.errMethod == v1alpha2.TRANSFORM_INPUT {
		return nil, s.err
	}
	return msg, nil
}

func (s SeldonMessageTestClient) TransformInput(host string, port int32, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	s.t.Logf("TransformInput %s %d", host, port)
	if s.errMethod != nil && *s.errMethod == v1alpha2.TRANSFORM_INPUT {
		return nil, s.err
	}
	return msg, nil
}

func (s SeldonMessageTestClient) Route(host string, port int32, msg payload.SeldonPayload) (int, error) {
	s.t.Logf("Route %s %d", host, port)
	return s.chosenRoute, nil
}

func (s SeldonMessageTestClient) Combine(host string, port int32, msgs []payload.SeldonPayload) (payload.SeldonPayload, error) {
	s.t.Logf("Combine %s %d", host, port)
	return msgs[0], nil
}

func (s SeldonMessageTestClient) TransformOutput(host string, port int32, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	s.t.Logf("TransformOutput %s %d", host, port)
	return msg, nil
}

func NewSeldonMessageTestClient(t *testing.T, chosenRoute int, errMethod *v1alpha2.PredictiveUnitMethod, err error) client.SeldonApiClient {
	client := SeldonMessageTestClient{
		t:           t,
		chosenRoute: chosenRoute,
		errMethod:   errMethod,
		err:         err,
	}
	return &client
}
