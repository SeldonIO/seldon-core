/*
Copyright 2022 Seldon Technologies Ltd.

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

package gateway

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/jarcoal/httpmock"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"

	v2 "github.com/seldonio/seldon-core/apis/go/v2/mlops/v2_dataplane"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/envoy/resources"
	kafka2 "github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/config"
	seldontracer "github.com/seldonio/seldon-core/scheduler/v2/pkg/tracing"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

func createTestV2ClientMockResponders(host string, port int, modelName string) {
	httpmock.RegisterResponder("POST", fmt.Sprintf("http://%s:%d/v2/models/%s/infer", host, port, modelName),
		httpmock.NewStringResponder(http.StatusOK, `{}`))
}

func TestRestRequest(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name string
		data []byte
	}
	tests := []test{
		{
			name: "smoke test",
			data: []byte{},
		},
		{
			name: "smoke empty test",
			data: []byte(""),
		},
		{
			name: "smoke json test",
			data: []byte("{}"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			kafkaServerConfig := InferenceServerConfig{
				Host:     "0.0.0.0",
				HttpPort: 1234,
				GrpcPort: 1235,
			}
			kafkaModelConfig := KafkaModelConfig{
				ModelName:   "foo",
				InputTopic:  "input",
				OutputTopic: "output",
			}
			createTestV2ClientMockResponders(kafkaServerConfig.Host, kafkaServerConfig.HttpPort, kafkaModelConfig.ModelName)
			logger := log.New()
			tp, err := seldontracer.NewTraceProvider("test", nil, logger)
			g.Expect(err).To(BeNil())
			config := &ManagerConfig{SeldonKafkaConfig: &config.KafkaConfig{}, Namespace: "default", InferenceServerConfig: &kafkaServerConfig, TraceProvider: tp, NumWorkers: 0}
			ic, err := NewInferKafkaHandler(logger, config, kafka.ConfigMap{}, kafka.ConfigMap{}, "dummy")
			g.Expect(err).To(BeNil())
			tn, err := kafka2.NewTopicNamer("default", "seldon")
			g.Expect(err).To(BeNil())
			iw, err := NewInferWorker(ic, logger, tp, tn)
			g.Expect(err).To(BeNil())
			err = iw.restRequest(context.Background(), &InferWork{modelName: "foo", msg: &kafka.Message{Value: test.data}}, false)
			g.Expect(err).To(BeNil())
			ic.Stop()
			g.Expect(httpmock.GetTotalCallCount()).To(Equal(1))
			g.Expect(ic.producer.Len()).To(Equal(1))
		})
	}
}

func TestProcessRequestRest(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name string
		data []byte
	}
	tests := []test{
		{
			name: "smoke test rest",
			data: []byte("{}"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			kafkaServerConfig := InferenceServerConfig{
				Host:     "0.0.0.0",
				HttpPort: 1234,
				GrpcPort: 1235,
			}
			kafkaModelConfig := KafkaModelConfig{
				ModelName:   "foo",
				InputTopic:  "input",
				OutputTopic: "output",
			}
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			createTestV2ClientMockResponders(kafkaServerConfig.Host, kafkaServerConfig.HttpPort, kafkaModelConfig.ModelName)
			logger := log.New()
			tp, err := seldontracer.NewTraceProvider("test", nil, logger)
			g.Expect(err).To(BeNil())
			config := &ManagerConfig{SeldonKafkaConfig: &config.KafkaConfig{}, Namespace: "default", InferenceServerConfig: &kafkaServerConfig, TraceProvider: tp, NumWorkers: 0}
			ic, err := NewInferKafkaHandler(logger, config, kafka.ConfigMap{}, kafka.ConfigMap{}, "dummy")
			g.Expect(err).To(BeNil())
			tn, err := kafka2.NewTopicNamer("default", "seldon")
			g.Expect(err).To(BeNil())
			iw, err := NewInferWorker(ic, logger, tp, tn)
			g.Expect(err).To(BeNil())
			err = iw.processRequest(context.Background(), &InferWork{modelName: "foo", msg: &kafka.Message{Value: test.data}})
			g.Expect(err).To(BeNil())
			ic.Stop()
			g.Eventually(httpmock.GetTotalCallCount).Should(Equal(1))
			g.Eventually(ic.producer.Len).Should(Equal(1))
		})
	}
}

const bufSize = 1024 * 1024

type mockGRPCMLServer struct {
	listener *bufconn.Listener
	server   *grpc.Server
	v2.UnimplementedGRPCInferenceServiceServer
	recv int
}

func (m *mockGRPCMLServer) setup() error {
	var err error
	m.listener = bufconn.Listen(bufSize)
	if err != nil {
		return err
	}
	opts := []grpc.ServerOption{}
	m.server = grpc.NewServer(opts...)
	v2.RegisterGRPCInferenceServiceServer(m.server, m)
	grpc_health_v1.RegisterHealthServer(m.server, health.NewServer())
	return nil
}

func (m *mockGRPCMLServer) start() error {
	return m.server.Serve(m.listener)
}

func (m *mockGRPCMLServer) stop() {
	_ = m.listener.Close()
	m.server.Stop()
}

func (m *mockGRPCMLServer) ModelInfer(ctx context.Context, r *v2.ModelInferRequest) (*v2.ModelInferResponse, error) {
	m.recv = m.recv + 1
	return &v2.ModelInferResponse{ModelName: r.ModelName, ModelVersion: r.ModelVersion}, nil
}

func createMLMockGrpcServer(g *GomegaWithT) *mockGRPCMLServer {
	mockMLServer := &mockGRPCMLServer{}
	err := mockMLServer.setup()
	g.Expect(err).To(BeNil())
	go func() {
		err := mockMLServer.start()
		g.Expect(err).To(BeNil())
	}()
	return mockMLServer
}

func createInferWorkerWithMockConn(
	grpcServer *mockGRPCMLServer,
	logger log.FieldLogger,
	serverConfig *InferenceServerConfig,
	modelConfig *KafkaModelConfig,
	g *WithT) (*InferKafkaHandler, *InferWorker) {
	conn, _ := grpc.DialContext(context.TODO(), "", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return grpcServer.listener.Dial()
	}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	tp, err := seldontracer.NewTraceProvider("test", nil, logger)
	g.Expect(err).To(BeNil())
	config := &ManagerConfig{SeldonKafkaConfig: &config.KafkaConfig{}, Namespace: "default", InferenceServerConfig: serverConfig, TraceProvider: tp, NumWorkers: 0}
	ic, err := NewInferKafkaHandler(logger, config, kafka.ConfigMap{}, kafka.ConfigMap{}, "dummy")
	g.Expect(err).To(BeNil())
	topicNamer, err := kafka2.NewTopicNamer("default", "seldon")
	g.Expect(err).To(BeNil())
	iw := &InferWorker{
		logger:     logger,
		grpcClient: v2.NewGRPCInferenceServiceClient(conn),
		httpClient: http.DefaultClient,
		consumer:   ic,
		tracer:     tp.GetTraceProvider().Tracer("test"),
		topicNamer: topicNamer,
	}
	return ic, iw
}

func creatMockServerHealthFunc(grpcServer *mockGRPCMLServer) func() bool {
	return func() bool {
		conn, _ := grpc.DialContext(context.TODO(), "", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return grpcServer.listener.Dial()
		}), grpc.WithTransportCredentials(insecure.NewCredentials()))
		client := grpc_health_v1.NewHealthClient(conn)
		hcr, err := client.Check(context.TODO(), &grpc_health_v1.HealthCheckRequest{})
		if err != nil || hcr.Status != grpc_health_v1.HealthCheckResponse_SERVING {
			return false
		}
		return true
	}
}

func TestProcessRequestGrpc(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name string
		req  *v2.ModelInferRequest
	}
	tests := []test{
		{
			name: "smoke grpc rest",
			req:  &v2.ModelInferRequest{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := log.New()
			t.Log("Start test", test.name)
			kafkaServerConfig := InferenceServerConfig{
				Host:     "0.0.0.0",
				HttpPort: 1234,
				GrpcPort: 1235,
			}
			kafkaModelConfig := KafkaModelConfig{
				ModelName:   "foo",
				InputTopic:  "input",
				OutputTopic: "output",
			}
			mockMLGrpcServer := createMLMockGrpcServer(g)
			defer mockMLGrpcServer.stop()
			ic, iw := createInferWorkerWithMockConn(mockMLGrpcServer, logger, &kafkaServerConfig, &kafkaModelConfig, g)
			defer ic.Stop()
			check := creatMockServerHealthFunc(mockMLGrpcServer)
			g.Eventually(check).Should(BeTrue())
			b, err := proto.Marshal(test.req)
			g.Expect(err).To(BeNil())
			err = iw.processRequest(context.Background(), &InferWork{modelName: "foo", msg: &kafka.Message{Value: b}})
			g.Expect(err).To(BeNil())
			g.Eventually(func() int { return mockMLGrpcServer.recv }).Should(Equal(1))
			g.Eventually(ic.producer.Len).Should(Equal(1))
			t.Log("End test", test.name)
		})
	}
}

func TestProcessRequest(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name      string
		job       *InferWork
		restCalls int
		grpcCalls int
		error     bool
	}
	getProtoBytes := func(res proto.Message) []byte {
		b, _ := proto.Marshal(res)
		return b
	}

	testRequest := &v2.ModelInferRequest{
		Inputs: []*v2.ModelInferRequest_InferInputTensor{
			{
				Name:     "out1",
				Datatype: "float",
				Shape:    []int64{1, 2, 3},
				Contents: &v2.InferTensorContents{
					IntContents: []int32{1, 2, 3},
				},
			},
		},
	}
	testResponse := &v2.ModelInferResponse{
		Outputs: []*v2.ModelInferResponse_InferOutputTensor{
			{
				Name:     "out1",
				Datatype: "float",
				Shape:    []int64{1, 2, 3},
				Contents: &v2.InferTensorContents{
					IntContents: []int32{1, 2, 3},
				},
			},
		},
	}
	tests := []test{
		{
			name: "empty request is assumed grpc",
			job: &InferWork{
				modelName: "foo",
				headers:   make(map[string]string),
				msg:       &kafka.Message{Value: []byte{}, Key: []byte{}},
			},
			grpcCalls: 1,
		},
		{
			name: "empty json request",
			job: &InferWork{
				modelName: "foo",
				headers:   make(map[string]string),
				msg:       &kafka.Message{Value: []byte("{}"), Key: []byte{}},
			},
			restCalls: 1,
		},
		{
			name: "json request",
			job: &InferWork{
				modelName: "foo",
				headers:   make(map[string]string),
				msg:       &kafka.Message{Value: []byte(`{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}`), Key: []byte{}},
			},
			restCalls: 1,
		},
		{
			name: "chain json request",
			job: &InferWork{
				modelName: "foo",
				headers:   make(map[string]string),
				msg:       &kafka.Message{Value: []byte(`{"model_name":"iris_1","model_version":"1","id":"903964e4-2419-41ce-b5d1-3ca0c8df9e0c","parameters":null,"outputs":[{"name":"predict","shape":[1],"datatype":"INT64","parameters":null,"data":[2]}]}`), Key: []byte{}},
			},
			restCalls: 1,
		},
		{
			name: "json request with header",
			job: &InferWork{
				modelName: "foo",
				headers:   map[string]string{HeaderKeyType: HeaderValueJsonReq},
				msg:       &kafka.Message{Value: []byte(`{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}`), Key: []byte{}},
			},
			restCalls: 1,
		},
		{
			name: "chain json request with header",
			job: &InferWork{
				modelName: "foo",
				headers:   map[string]string{HeaderKeyType: HeaderValueJsonRes},
				msg:       &kafka.Message{Value: []byte(`{"model_name":"iris_1","model_version":"1","id":"903964e4-2419-41ce-b5d1-3ca0c8df9e0c","parameters":null,"outputs":[{"name":"predict","shape":[1],"datatype":"INT64","parameters":null,"data":[2]}]}`), Key: []byte{}},
			},
			restCalls: 1,
		},
		{
			name: "grpc request without header",
			job: &InferWork{
				modelName: "foo",
				headers:   make(map[string]string),
				msg:       &kafka.Message{Value: getProtoBytes(testRequest), Key: []byte{}},
			},
			grpcCalls: 1,
		},
		{
			name: "grpc request with header",
			job: &InferWork{
				modelName: "foo",
				headers:   map[string]string{HeaderKeyType: HeaderValueProtoReq},
				msg:       &kafka.Message{Value: getProtoBytes(testRequest), Key: []byte{}},
			},
			grpcCalls: 1,
		},
		{
			name: "chained grpc request without header",
			job: &InferWork{
				modelName: "foo",
				headers:   make(map[string]string),
				msg:       &kafka.Message{Value: getProtoBytes(testResponse), Key: []byte{}},
			},
			grpcCalls: 1,
		},
		{
			name: "chained grpc request with header",
			job: &InferWork{
				modelName: "foo",
				headers:   map[string]string{HeaderKeyType: HeaderValueProtoRes},
				msg:       &kafka.Message{Value: getProtoBytes(testResponse), Key: []byte{}},
			},
			grpcCalls: 1,
		},
		{
			name: "json request with proto request header",
			job: &InferWork{
				modelName: "foo",
				headers:   map[string]string{HeaderKeyType: HeaderValueProtoReq},
				msg:       &kafka.Message{Value: []byte(`{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}`), Key: []byte{}},
			},
			error: true,
		},
		{
			name: "json request with proto response header",
			job: &InferWork{
				modelName: "foo",
				headers:   map[string]string{HeaderKeyType: HeaderValueProtoRes},
				msg:       &kafka.Message{Value: []byte(`{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}`), Key: []byte{}},
			},
			error: true,
		},
		{
			name: "grpc request with json header treated as json", //TODO maybe fail in this case as it will fail at server
			job: &InferWork{
				modelName: "foo",
				headers:   map[string]string{HeaderKeyType: HeaderValueJsonReq},
				msg:       &kafka.Message{Value: getProtoBytes(testRequest), Key: []byte{}},
			},
			restCalls: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := log.New()
			logger.Infof("Start test %s", test.name)
			t.Log("Start test", test.name)
			kafkaServerConfig := InferenceServerConfig{
				Host:     "0.0.0.0",
				HttpPort: 1234,
				GrpcPort: 1235,
			}
			kafkaModelConfig := KafkaModelConfig{
				ModelName:   "foo",
				InputTopic:  "input",
				OutputTopic: "output",
			}
			mockMLGrpcServer := createMLMockGrpcServer(g)
			defer mockMLGrpcServer.stop()
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			createTestV2ClientMockResponders(kafkaServerConfig.Host, kafkaServerConfig.HttpPort, kafkaModelConfig.ModelName)
			ic, iw := createInferWorkerWithMockConn(mockMLGrpcServer, logger, &kafkaServerConfig, &kafkaModelConfig, g)
			defer ic.Stop()
			check := creatMockServerHealthFunc(mockMLGrpcServer)
			g.Eventually(check).Should(BeTrue())
			err := iw.processRequest(context.Background(), test.job)
			if test.error {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				g.Eventually(httpmock.GetTotalCallCount).Should(Equal(test.restCalls))
				g.Eventually(func() int { return mockMLGrpcServer.recv }).Should(Equal(test.grpcCalls))
				g.Eventually(ic.producer.Len).Should(Equal(1))
			}
			t.Log("End test", test.name)
		})
	}
}

func TestAddMetadataToOutgoingContext(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name            string
		ctx             context.Context
		job             *InferWork
		expectedHeaders map[string][]string
	}

	tests := []test{
		{
			name:            "ignore xseldon-route header",
			ctx:             metadata.NewIncomingContext(context.TODO(), metadata.New(map[string]string{})),
			job:             &InferWork{modelName: "foo", headers: map[string]string{resources.SeldonRouteHeader: ":a:"}},
			expectedHeaders: map[string][]string{resources.SeldonModelHeader: {"foo"}},
		},
		{
			name:            "pass x-request-id header",
			ctx:             metadata.NewIncomingContext(context.TODO(), metadata.New(map[string]string{})),
			job:             &InferWork{modelName: "foo", headers: map[string]string{util.RequestIdHeader: "1234"}},
			expectedHeaders: map[string][]string{resources.SeldonModelHeader: {"foo"}, util.RequestIdHeader: {"1234"}},
		},
		{
			name:            "pass custom header",
			ctx:             metadata.NewIncomingContext(context.TODO(), metadata.New(map[string]string{})),
			job:             &InferWork{modelName: "foo", headers: map[string]string{"x-myheader": "1234"}},
			expectedHeaders: map[string][]string{resources.SeldonModelHeader: {"foo"}, "x-myheader": {"1234"}},
		},
		{
			name:            "ignore non x- prefix headers",
			ctx:             metadata.NewIncomingContext(context.TODO(), metadata.New(map[string]string{})),
			job:             &InferWork{modelName: "foo", headers: map[string]string{"myheader": "1234"}},
			expectedHeaders: map[string][]string{resources.SeldonModelHeader: {"foo"}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := log.New()
			ctx := addMetadataToOutgoingContext(test.ctx, test.job, logger)
			md, found := metadata.FromOutgoingContext(ctx)
			g.Expect(found).To(BeTrue())
			for k, v := range md {
				vExpected, ok := test.expectedHeaders[k]
				g.Expect(ok).To(BeTrue())
				g.Expect(v).To(Equal(vExpected))
			}
		})
	}
}
