package tensorflow

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/executor/api/client"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/proto/tensorflow/serving"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	status "google.golang.org/grpc/status"
)

const (
	TestMetaDataKey  = "foo"
	TestMetaDataVal  = "bar"
	TestModelVersion = int64(1)
)

type TestTensorflowClient struct {
	t *testing.T
}

func (s TestTensorflowClient) IsGrpc() bool {
	return true
}

func NewTestTensorflowClient(t *testing.T) client.SeldonApiClient {
	client := TestTensorflowClient{
		t: t,
	}
	return &client
}

func (s TestTensorflowClient) Status(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	st := serving.GetModelStatusResponse{
		ModelVersionStatus: []*serving.ModelVersionStatus{
			&serving.ModelVersionStatus{
				Version: TestModelVersion,
			},
		},
	}
	sm := payload.ProtoPayload{Msg: &st}
	return &sm, nil
}

func (s TestTensorflowClient) Metadata(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	g := NewGomegaWithT(s.t)
	md, ok := metadata.FromIncomingContext(ctx)
	g.Expect(ok).NotTo(BeNil())
	g.Expect(md.Get(TestMetaDataKey)[0]).To(Equal(TestMetaDataVal))
	pm := msg.GetPayload().(*serving.GetModelMetadataRequest)
	st := serving.GetModelMetadataResponse{
		ModelSpec: pm.ModelSpec,
	}
	sm := payload.ProtoPayload{Msg: &st}
	return &sm, nil
}

func (s TestTensorflowClient) ModelMetadata(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.ModelMetadata, error) {
	return payload.ModelMetadata{}, status.Errorf(codes.Unimplemented, "ModelMetadata not implemented")
}

func (s TestTensorflowClient) Chain(ctx context.Context, modelName string, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	return msg, nil
}

func (s TestTensorflowClient) Unmarshall(msg []byte, contentType string) (payload.SeldonPayload, error) {
	reqPayload := payload.BytesPayload{Msg: msg, ContentType: contentType}
	return &reqPayload, nil
}

func (s TestTensorflowClient) Marshall(out io.Writer, msg payload.SeldonPayload) error {
	panic("")
}

func (s TestTensorflowClient) CreateErrorPayload(err error) payload.SeldonPayload {
	respFailed := proto.SeldonMessage{Status: &proto.Status{Code: http.StatusInternalServerError, Info: err.Error()}}
	res := payload.ProtoPayload{Msg: &respFailed}
	return &res
}

func (s TestTensorflowClient) Predict(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	g := NewGomegaWithT(s.t)
	md, ok := metadata.FromIncomingContext(ctx)
	g.Expect(ok).NotTo(BeNil())
	g.Expect(md.Get(TestMetaDataKey)[0]).To(Equal(TestMetaDataVal))
	pm := msg.GetPayload().(*serving.PredictRequest)
	pr := serving.PredictResponse{
		Outputs: pm.Inputs,
	}
	sm := payload.ProtoPayload{Msg: &pr}
	return &sm, nil
}

func (s TestTensorflowClient) TransformInput(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	panic("")
}

func (s TestTensorflowClient) Route(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (int, error) {
	panic("")
}

func (s TestTensorflowClient) Combine(ctx context.Context, modelName string, host string, port int32, msgs []payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	panic("")
}

func (s TestTensorflowClient) TransformOutput(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	panic("")
}

func (s TestTensorflowClient) Feedback(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	panic("not implemented")
}
func TestPredict(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	model := v1.MODEL
	p := v1.PredictorSpec{
		Name: "p",
		Graph: v1.PredictiveUnit{
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1.REST,
			},
		},
	}
	url, _ := url.Parse("http://localhost")
	server := NewGrpcTensorflowServer(&p, NewTestTensorflowClient(t), url, "default")

	var sm serving.PredictRequest
	var data = `{"model_spec":{"name":"half_plus_two"},"inputs":{"x":{"dtype": 1, "tensor_shape": {"dim":[{"size": 3}]}, "floatVal" : [1.0, 2.0, 3.0]}}}`
	err := jsonpb.UnmarshalString(data, &sm)
	g.Expect(err).Should(BeNil())

	ctx := context.Background()
	ctx = metadata.NewIncomingContext(ctx, metadata.New(map[string]string{TestMetaDataKey: TestMetaDataVal}))
	res, err := server.Predict(ctx, &sm)
	g.Expect(err).To(BeNil())
	g.Expect(res.Outputs["x"].FloatVal[0]).Should(Equal(float32(1.0)))
	g.Expect(res.Outputs["x"].FloatVal[1]).Should(Equal(float32(2.0)))
}

func TestGetModelStatus(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	model := v1.MODEL
	p := v1.PredictorSpec{
		Name: "p",
		Graph: v1.PredictiveUnit{
			Name: "model",
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1.REST,
			},
		},
	}
	url, _ := url.Parse("http://localhost")
	server := NewGrpcTensorflowServer(&p, NewTestTensorflowClient(t), url, "default")

	var sm serving.GetModelStatusRequest
	var data = `{"model_spec":{"name":"model"}}`
	err := jsonpb.UnmarshalString(data, &sm)
	g.Expect(err).Should(BeNil())

	ctx := context.Background()
	ctx = metadata.NewIncomingContext(ctx, metadata.New(map[string]string{TestMetaDataKey: TestMetaDataVal}))
	res, err := server.GetModelStatus(ctx, &sm)
	g.Expect(err).To(BeNil())
	g.Expect(res.ModelVersionStatus[0].Version).To(Equal(TestModelVersion))

}

func TestGetModelMetadata(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	const modelName = "model"
	model := v1.MODEL
	p := v1.PredictorSpec{
		Name: "p",
		Graph: v1.PredictiveUnit{
			Name: modelName,
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1.REST,
			},
		},
	}
	url, _ := url.Parse("http://localhost")
	server := NewGrpcTensorflowServer(&p, NewTestTensorflowClient(t), url, "default")

	var sm serving.GetModelMetadataRequest
	var data = `{"model_spec":{"name":"model"},"metadata_field":["signature_def"]}`
	err := jsonpb.UnmarshalString(data, &sm)
	g.Expect(err).Should(BeNil())

	ctx := context.Background()
	ctx = metadata.NewIncomingContext(ctx, metadata.New(map[string]string{TestMetaDataKey: TestMetaDataVal}))
	res, err := server.GetModelMetadata(ctx, &sm)
	g.Expect(err).To(BeNil())
	g.Expect(res.ModelSpec.Name).To(Equal(modelName))

}
