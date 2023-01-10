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

package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	v2 "github.com/seldonio/seldon-core/apis/go/v2/mlops/v2_dataplane"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MLServerModelState string

const (
	MLServerModelState_UNKNOWN     MLServerModelState = "UNKNOWN"
	MLServerModelState_READY       MLServerModelState = "READY"
	MLServerModelState_UNAVAILABLE MLServerModelState = "UNAVAILABLE"
	MLServerModelState_LOADING     MLServerModelState = "LOADING"
	MLServerModelState_UNLOADING   MLServerModelState = "UNLOADING"
)

const (
	// we define all communication error into one bucket
	// TODO: separate out the different comm issues (e.g. DNS vs Connection refused etc.)
	V2CommunicationErrCode = -100
	// i.e invalid method etc.
	V2RequestErrCode = -200
)

type MLServerModelInfo struct {
	Name  string
	State MLServerModelState
}

type V2Client struct {
	host       string
	httpPort   int
	httpClient *http.Client
	grpcPort   int
	grpcClient v2.GRPCInferenceServiceClient
	logger     log.FieldLogger
	isGrpc     bool
}

// Error wrapper with client and server errors + error code
// errCode should have the standard http error codes (for server)
// and client communication error codes (defined above)
type V2Err struct {
	isGrpc bool
	err    error
	// one bucket for http/grpc status code and client codes (for error)
	errCode int
}

func (v2err *V2Err) Error() string {
	if v2err != nil {
		return v2err.err.Error()
	}
	return ""
}

func (v *V2Err) IsNotFound() bool {
	if v.isGrpc {
		return v.errCode == int(codes.NotFound)
	} else {
		// assumes http
		return v.errCode == http.StatusNotFound
	}
}

type V2ServerError struct {
	Error string `json:"error"`
}

var ErrV2BadRequest = errors.New("V2 Bad Request")
var ErrServerNotReady = errors.New("Server not ready")

func getV2GrpcConnection(host string, plainTxtPort int) (*grpc.ClientConn, error) {
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(util.GrpcRetryBackoffMillisecs * time.Millisecond)),
		grpc_retry.WithMax(util.GrpcRetryMaxCount),
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(retryOpts...)),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(grpc_retry.UnaryClientInterceptor(retryOpts...), otelgrpc.UnaryClientInterceptor())),
	}
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, plainTxtPort), opts...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func createV2ControlPlaneClient(host string, port int) (v2.GRPCInferenceServiceClient, error) {
	conn, err := getV2GrpcConnection(host, port)
	if err != nil {
		// TODO: this could fail in later iterations, so close earlier connections
		conn.Close()
		return nil, err
	}

	client := v2.NewGRPCInferenceServiceClient(conn)
	return client, nil
}

func NewV2Client(host string, port int, logger log.FieldLogger, isGrpc bool) *V2Client {
	logger.Infof("V2 Inference Server %s:%d", host, port)

	if isGrpc {
		grpcClient, err := createV2ControlPlaneClient(host, port)
		if err != nil {
			return nil
		}

		return &V2Client{
			host:       host,
			grpcPort:   port,
			grpcClient: grpcClient,
			logger:     logger.WithField("Source", "V2InferenceServerClientGrpc"),
			isGrpc:     isGrpc,
		}
	} else {
		netTransport := &http.Transport{
			MaxIdleConns:        maxIdleConnsHTTP,
			MaxIdleConnsPerHost: maxIdleConnsPerHostHTTP,
			DisableKeepAlives:   disableKeepAlivesHTTP,
			MaxConnsPerHost:     maxConnsPerHostHTTP,
			IdleConnTimeout:     idleConnTimeoutSeconds * time.Second,
		}
		netClient := &http.Client{
			Timeout:   time.Second * defaultTimeoutSeconds,
			Transport: netTransport,
		}

		return &V2Client{
			host:       host,
			httpPort:   port,
			httpClient: netClient,
			logger:     logger.WithField("Source", "V2InferenceServerClientHttp"),
			isGrpc:     isGrpc,
		}
	}
}

func (v *V2Client) getUrl(path string) *url.URL {
	return &url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(v.host, strconv.Itoa(v.httpPort)),
		Path:   path,
	}
}

func (v *V2Client) call(path string) *V2Err {
	v2Url := v.getUrl(path)
	req, err := http.NewRequest("POST", v2Url.String(), bytes.NewBuffer([]byte{}))
	if err != nil {
		return &V2Err{
			isGrpc:  false,
			err:     err,
			errCode: V2RequestErrCode,
		}
	}
	response, err := v.httpClient.Do(req)
	if err != nil {
		return &V2Err{
			isGrpc:  false,
			err:     err,
			errCode: V2CommunicationErrCode,
		}
	}
	b, err := io.ReadAll(response.Body)
	if err != nil {
		return &V2Err{
			isGrpc:  false,
			err:     err,
			errCode: response.StatusCode,
		}
	}
	err = response.Body.Close()
	if err != nil {
		return &V2Err{
			isGrpc:  false,
			err:     err,
			errCode: response.StatusCode,
		}
	}
	v.logger.Infof("v2 server response: %s", b)
	if response.StatusCode != http.StatusOK {
		if response.StatusCode == http.StatusBadRequest {
			v2Error := V2ServerError{}
			err := json.Unmarshal(b, &v2Error)
			if err != nil {
				return &V2Err{
					isGrpc:  false,
					err:     err,
					errCode: response.StatusCode,
				}
			}
			return &V2Err{
				isGrpc:  false,
				err:     fmt.Errorf("%s. %w", v2Error.Error, ErrV2BadRequest),
				errCode: response.StatusCode,
			}
		} else {
			return &V2Err{
				isGrpc:  false,
				err:     fmt.Errorf("V2 server error: %s", b),
				errCode: response.StatusCode,
			}
		}
	}
	return nil
}

func (v *V2Client) LoadModel(name string) *V2Err {
	if v.isGrpc {
		return v.loadModelGrpc(name)
	} else {
		return v.loadModelHttp(name)
	}
}

func (v *V2Client) loadModelHttp(name string) *V2Err {
	path := fmt.Sprintf("v2/repository/models/%s/load", name)
	v.logger.Infof("Load request: %s", path)
	return v.call(path)
}

func (v *V2Client) loadModelGrpc(name string) *V2Err {
	ctx := context.Background()

	req := &v2.RepositoryModelLoadRequest{
		ModelName: name,
	}

	_, err := v.grpcClient.RepositoryModelLoad(ctx, req)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			errCode := e.Code()
			return &V2Err{
				err:     err,
				errCode: int(errCode),
				isGrpc:  true,
			}
		}
		return &V2Err{
			err:     err,
			errCode: V2CommunicationErrCode,
			isGrpc:  true,
		}

	}
	return nil
}

func (v *V2Client) UnloadModel(name string) *V2Err {
	if v.isGrpc {
		return v.unloadModelGrpc(name)
	} else {
		return v.unloadModelHttp(name)
	}
}

func (v *V2Client) unloadModelHttp(name string) *V2Err {
	path := fmt.Sprintf("v2/repository/models/%s/unload", name)
	v.logger.Infof("Unload request: %s", path)
	return v.call(path)
}

func (v *V2Client) unloadModelGrpc(name string) *V2Err {
	ctx := context.Background()

	req := &v2.RepositoryModelUnloadRequest{
		ModelName: name,
	}

	_, err := v.grpcClient.RepositoryModelUnload(ctx, req)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			errCode := e.Code()
			return &V2Err{
				err:     err,
				errCode: int(errCode),
				isGrpc:  true,
			}
		}
		return &V2Err{
			err:     err,
			errCode: V2CommunicationErrCode,
			isGrpc:  true,
		}
	}
	return nil
}

func (v *V2Client) Ready() error {
	var ready bool
	var err error
	if v.isGrpc {
		ready, err = v.readyGrpc()
	} else {
		ready, err = v.readyHttp()
	}
	if err != nil {
		v.logger.WithError(err).Debugf("Server ready check failed on error")
		return err
	}
	v.logger.Debugf("Server ready check returned with value %v", ready)
	if ready {
		return nil
	} else {
		return ErrServerNotReady
	}
}

func (v *V2Client) readyHttp() (bool, error) {
	res, err := http.Get(v.getUrl("v2/health/ready").String())
	if err != nil {
		return false, err
	}
	if res.StatusCode == http.StatusOK {
		return true, nil
	} else {
		return false, nil
	}
}

func (v *V2Client) readyGrpc() (bool, error) {
	ctx := context.Background()
	req := &v2.ServerReadyRequest{}

	res, err := v.grpcClient.ServerReady(ctx, req)
	if err != nil {
		return false, err
	}
	return res.Ready, nil
}

func (v *V2Client) GetModels() ([]MLServerModelInfo, error) {
	if v.isGrpc {
		return v.getModelsGrpc()
	} else {
		v.logger.Warnf("Http GetModels not available returning empty list")
		return []MLServerModelInfo{}, nil
	}
}

func (v *V2Client) getModelsGrpc() ([]MLServerModelInfo, error) {
	var models []MLServerModelInfo
	ctx := context.Background()
	req := &v2.RepositoryIndexRequest{}

	res, err := v.grpcClient.RepositoryIndex(ctx, req)
	if err != nil {
		return nil, err
	}
	for _, modelRes := range res.Models {
		if modelRes.Name == "" {
			// nothing to do for empty model
			// TODO: why mlserver returns back empty string model?
			continue
		}
		models = append(
			models, MLServerModelInfo{Name: modelRes.Name, State: MLServerModelState(modelRes.State)})
	}
	return models, nil
}
