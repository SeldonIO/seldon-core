package pipeline

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	v2 "github.com/seldonio/seldon-core/scheduler/apis/mlops/v2_dataplane"
	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/resources"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

type fakePipelineInferer struct {
	err  error
	data []byte
}

func (f *fakePipelineInferer) Infer(ctx context.Context, resourceName string, isModel bool, data []byte) ([]byte, error) {
	return f.data, f.err
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func waitForServer(port int) {
	backoff := 50 * time.Millisecond

	for i := 0; i < 10; i++ {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf(":%d", port), 1*time.Second)
		if err != nil {
			time.Sleep(backoff)
			continue
		}
		err = conn.Close()
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	log.Fatalf("Server on port %d not up after 10 attempts", port)
}

func TestHttpServer(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name       string
		path       string
		header     string
		req        string
		res        *v2.ModelInferResponse
		statusCode int
	}
	tests := []test{
		{
			name:   "ok",
			path:   "/v2/models/foo/infer",
			header: "foo",
			req:    `{"inputs":[{"name":"input1","datatype":"BOOL","shape":[5],"data":[true,false,true,false,true]}]}`,
			res: &v2.ModelInferResponse{
				ModelName:    "model",
				ModelVersion: "1",
				Id:           "1234",
				Outputs: []*v2.ModelInferResponse_InferOutputTensor{
					{
						Name:     "t1",
						Datatype: tyBool,
						Shape:    []int64{2},
						Contents: &v2.InferTensorContents{
							BoolContents: []bool{true, false},
						},
					},
				},
			},
			statusCode: http.StatusOK,
		},
		{
			name:       "wrong path",
			path:       "/foo",
			header:     "foo",
			req:        `{"inputs":[{"name":"input1","datatype":"BOOL","shape":[5],"data":[true,false,true,false,true]}]}`,
			statusCode: http.StatusNotFound,
		},
		{
			name:       "bad header",
			path:       "/v2/models/foo/infer",
			header:     "foo.bar.bar",
			req:        `{"inputs":[{"name":"input1","datatype":"BOOL","shape":[5],"data":[true,false,true,false,true]}]}`,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "bad request",
			path:       "/v2/models/foo/infer",
			header:     "foo",
			req:        ``,
			statusCode: http.StatusBadRequest,
		},
	}

	port, err := getFreePort()
	g.Expect(err).To(BeNil())
	mockInferer := &fakePipelineInferer{
		err:  nil,
		data: []byte("result"),
	}
	httpServer := NewGatewayHttpServer(port, logrus.New(), nil, mockInferer, fakeMetricsHandler{})
	go func() {
		err := httpServer.Start()
		g.Expect(err).To(Equal(http.ErrServerClosed))
	}()
	waitForServer(port)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b, err := proto.Marshal(test.res)
			g.Expect(err).To(BeNil())
			mockInferer := &fakePipelineInferer{
				err:  nil,
				data: b,
			}
			httpServer.gateway = mockInferer
			inferV2Path := test.path
			url := "http://localhost:" + strconv.Itoa(port) + inferV2Path
			r := strings.NewReader(test.req)
			req, err := http.NewRequest(http.MethodPost, url, r)
			g.Expect(err).To(BeNil())
			req.Header.Set(resources.SeldonModelHeader, test.header)
			req.Header.Set("contentType", "application/json")
			resp, err := http.DefaultClient.Do(req)
			g.Expect(err).To(BeNil())
			g.Expect(resp.StatusCode).To(Equal(test.statusCode))
			defer resp.Body.Close()
		})
	}
	err = httpServer.Stop()
	g.Expect(err).To(BeNil())
}
