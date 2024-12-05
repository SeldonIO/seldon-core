/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package pipeline

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"

	v2 "github.com/seldonio/seldon-core/apis/go/v2/mlops/v2_dataplane"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/internal/testing_utils"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

type fakePipelineInferer struct {
	err          error
	data         []byte
	key          string
	isPayloadErr bool
	errorModel   string
}

func (f *fakePipelineInferer) Infer(ctx context.Context, resourceName string, isModel bool, data []byte, headers []kafka.Header, requestId string) (*Request, error) {
	if f.err != nil {
		return nil, f.err
	} else {
		return &Request{key: f.key, response: f.data, isError: f.isPayloadErr, errorModel: f.errorModel}, nil
	}
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
		name         string
		path         string
		header       string
		req          string
		res          *v2.ModelInferResponse
		errRes       []byte
		errorModel   string
		isPayloadErr bool
		statusCode   int
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
			name:         "payload error",
			path:         "/v2/models/foo/infer",
			header:       "foo",
			req:          `{"inputs":[{"name":"input1","datatype":"BOOL","shape":[500],"data":[true,false,true,false,true]}]}`,
			res:          &v2.ModelInferResponse{},
			isPayloadErr: true,
			errorModel:   "foo",
			errRes:       []byte("bad call"),
			statusCode:   http.StatusBadRequest,
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

	testRequestId := "test-id"
	port, err := testing_utils.GetFreePortForTest()
	g.Expect(err).To(BeNil())
	mockInferer := &fakePipelineInferer{
		err:  nil,
		data: []byte("result"),
		key:  testRequestId,
	}
	httpServer := NewGatewayHttpServer(port, logrus.New(), mockInferer, fakePipelineMetricsHandler{}, &util.TLSOptions{}, nil)
	go func() {
		err := httpServer.Start()
		g.Expect(err).To(Equal(http.ErrServerClosed))
	}()
	waitForServer(port)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var b []byte
			if test.errRes != nil {
				b = test.errRes
			} else {
				b, err = proto.Marshal(test.res)
				g.Expect(err).To(BeNil())
			}
			mockInferer := &fakePipelineInferer{
				err:          nil,
				data:         b,
				key:          testRequestId,
				isPayloadErr: test.isPayloadErr,
				errorModel:   test.errorModel,
			}
			httpServer.gateway = mockInferer
			inferV2Path := test.path
			url := "http://localhost:" + strconv.Itoa(port) + inferV2Path
			r := strings.NewReader(test.req)
			req, err := http.NewRequest(http.MethodPost, url, r)
			g.Expect(err).To(BeNil())
			req.Header.Set(util.SeldonModelHeader, test.header)
			req.Header.Set("contentType", "application/json")
			resp, err := http.DefaultClient.Do(req)
			g.Expect(err).To(BeNil())
			g.Expect(resp.StatusCode).To(Equal(test.statusCode))
			if resp.StatusCode == http.StatusOK {
				g.Expect(resp.Header.Get(util.RequestIdHeader)).ToNot(BeNil())
				g.Expect(resp.Header.Get(util.RequestIdHeader)).To(Equal(testRequestId))
			}
			if test.res != nil {
				bResp, err := io.ReadAll(resp.Body)
				g.Expect(err).To(BeNil())
				if resp.StatusCode == http.StatusOK {
					b, err := ConvertV2ResponseBytesToJson(b)
					g.Expect(err).To(BeNil())
					g.Expect(bResp).To(Equal(b))
				} else {
					g.Expect(bResp).To(Equal(createResponseErrorPayload(test.errorModel, test.errRes)))
				}
			}
			defer resp.Body.Close()
		})
	}
	err = httpServer.Stop()
	g.Expect(err).To(BeNil())
}
