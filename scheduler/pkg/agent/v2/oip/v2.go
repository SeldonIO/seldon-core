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

package oip

import (
	"bytes"
	"context"
	"encoding/json"
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
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/status"
)

type V2Client struct {
	host       string
	httpPort   int
	httpClient *http.Client
	grpcPort   int
	grpcClient v2.GRPCInferenceServiceClient
	logger     log.FieldLogger
	isGrpc     bool
}

func GetV2GrpcConnection(host string, plainTxtPort int) (*grpc.ClientConn, error) {
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(util.GrpcRetryBackoffMillisecs * time.Millisecond)),
		grpc_retry.WithMax(util.GrpcRetryMaxCount),
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(util.GrpcMaxMsgSizeBytes), grpc.MaxCallSendMsgSize(util.GrpcMaxMsgSizeBytes)),
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
	conn, err := GetV2GrpcConnection(host, port)
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
			MaxIdleConns:        util.MaxIdleConnsHTTP,
			MaxIdleConnsPerHost: util.MaxIdleConnsPerHostHTTP,
			DisableKeepAlives:   util.DisableKeepAlivesHTTP,
			MaxConnsPerHost:     util.MaxConnsPerHostHTTP,
			IdleConnTimeout:     util.IdleConnTimeoutSeconds * time.Second,
		}
		netClient := &http.Client{
			Timeout:   time.Second * util.DefaultTimeoutSeconds,
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

func (v *V2Client) call(path string) *interfaces.V2Err {
	v2Url := v.getUrl(path)
	req, err := http.NewRequest("POST", v2Url.String(), bytes.NewBuffer([]byte{}))
	if err != nil {
		return &interfaces.V2Err{
			IsGrpc:  false,
			Err:     err,
			ErrCode: interfaces.V2RequestErrCode,
		}
	}
	response, err := v.httpClient.Do(req)
	if err != nil {
		return &interfaces.V2Err{
			IsGrpc:  false,
			Err:     err,
			ErrCode: interfaces.V2CommunicationErrCode,
		}
	}
	b, err := io.ReadAll(response.Body)
	if err != nil {
		return &interfaces.V2Err{
			IsGrpc:  false,
			Err:     err,
			ErrCode: response.StatusCode,
		}
	}
	err = response.Body.Close()
	if err != nil {
		return &interfaces.V2Err{
			IsGrpc:  false,
			Err:     err,
			ErrCode: response.StatusCode,
		}
	}
	v.logger.Infof("v2 server response: %s", b)
	if response.StatusCode != http.StatusOK {
		if response.StatusCode == http.StatusBadRequest {
			v2Error := interfaces.V2ServerError{}
			err := json.Unmarshal(b, &v2Error)
			if err != nil {
				return &interfaces.V2Err{
					IsGrpc:  false,
					Err:     err,
					ErrCode: response.StatusCode,
				}
			}
			return &interfaces.V2Err{
				IsGrpc:  false,
				Err:     fmt.Errorf("%s. %w", v2Error.Error, interfaces.ErrV2BadRequest),
				ErrCode: response.StatusCode,
			}
		} else {
			return &interfaces.V2Err{
				IsGrpc:  false,
				Err:     fmt.Errorf("V2 server error: %s", b),
				ErrCode: response.StatusCode,
			}
		}
	}
	return nil
}

func (v *V2Client) LoadModel(name string) *interfaces.V2Err {
	if v.isGrpc {
		return v.loadModelGrpc(name)
	} else {
		return v.loadModelHttp(name)
	}
}

func (v *V2Client) loadModelHttp(name string) *interfaces.V2Err {
	path := fmt.Sprintf("v2/repository/models/%s/load", name)
	v.logger.Infof("Load request: %s", path)
	return v.call(path)
}

func (v *V2Client) loadModelGrpc(name string) *interfaces.V2Err {
	ctx := context.Background()

	req := &v2.RepositoryModelLoadRequest{
		ModelName: name,
	}

	_, err := v.grpcClient.RepositoryModelLoad(ctx, req)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			errCode := e.Code()
			return &interfaces.V2Err{
				Err:     err,
				ErrCode: int(errCode),
				IsGrpc:  true,
			}
		}
		return &interfaces.V2Err{
			Err:     err,
			ErrCode: interfaces.V2CommunicationErrCode,
			IsGrpc:  true,
		}

	}
	return nil
}

func (v *V2Client) UnloadModel(name string) *interfaces.V2Err {
	if v.isGrpc {
		return v.unloadModelGrpc(name)
	} else {
		return v.unloadModelHttp(name)
	}
}

func (v *V2Client) unloadModelHttp(name string) *interfaces.V2Err {
	path := fmt.Sprintf("v2/repository/models/%s/unload", name)
	v.logger.Infof("Unload request: %s", path)
	return v.call(path)
}

func (v *V2Client) unloadModelGrpc(name string) *interfaces.V2Err {
	ctx := context.Background()

	req := &v2.RepositoryModelUnloadRequest{
		ModelName: name,
	}

	_, err := v.grpcClient.RepositoryModelUnload(ctx, req)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			errCode := e.Code()
			return &interfaces.V2Err{
				Err:     err,
				ErrCode: int(errCode),
				IsGrpc:  true,
			}
		}
		return &interfaces.V2Err{
			Err:     err,
			ErrCode: interfaces.V2CommunicationErrCode,
			IsGrpc:  true,
		}
	}
	return nil
}

func (v *V2Client) Live() error {
	var ready bool
	var err error
	if v.isGrpc {
		ready, err = v.liveGrpc()
	} else {
		ready, err = v.liveHttp()
	}
	if err != nil {
		v.logger.WithError(err).Debugf("Server live check failed on error")
		return err
	}
	if ready {
		return nil
	} else {
		return interfaces.ErrServerNotReady
	}
}

func (v *V2Client) liveHttp() (bool, error) {
	res, err := http.Get(v.getUrl("v2/health/live").String())
	if err != nil {
		return false, err
	}
	if res.StatusCode == http.StatusOK {
		return true, nil
	} else {
		return false, nil
	}
}

func (v *V2Client) liveGrpc() (bool, error) {
	ctx := context.Background()
	req := &v2.ServerLiveRequest{}

	res, err := v.grpcClient.ServerLive(ctx, req)
	if err != nil {
		return false, err
	}
	return res.Live, nil
}

func (v *V2Client) GetModels() ([]interfaces.ServerModelInfo, error) {
	if v.isGrpc {
		return v.getModelsGrpc()
	} else {
		v.logger.Warnf("Http GetModels not available returning empty list")
		return []interfaces.ServerModelInfo{}, nil
	}
}

func (v *V2Client) getModelsGrpc() ([]interfaces.ServerModelInfo, error) {
	var models []interfaces.ServerModelInfo
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
			models, interfaces.ServerModelInfo{
				Name:  modelRes.Name,
				State: interfaces.ServerModelState(modelRes.State)})
	}
	return models, nil
}
