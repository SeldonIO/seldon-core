/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package oip

import (
	"context"
	"fmt"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	v2 "github.com/seldonio/seldon-core/apis/go/v2/mlops/v2_dataplane"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

type V2Config struct {
	Host                         string
	GRPCPort                     int
	GRPCRetryBackoff             time.Duration
	GRPRetryMaxCount             uint
	GRPCMaxMsgSizeBytes          int
	GRPCModelServerLoadTimeout   time.Duration
	GRPCModelServerUnloadTimeout time.Duration
	GRPCControlPlaneTimeout      time.Duration
}

type V2Client struct {
	grpcClient v2.GRPCInferenceServiceClient
	v2Config   V2Config
	logger     log.FieldLogger
}

func CreateV2GrpcConnection(v2Config V2Config) (*grpc.ClientConn, error) {
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(v2Config.GRPCRetryBackoff)),
		grpc_retry.WithMax(v2Config.GRPRetryMaxCount),
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(v2Config.GRPCMaxMsgSizeBytes), grpc.MaxCallSendMsgSize(v2Config.GRPCMaxMsgSizeBytes)),
		grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(retryOpts...)),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOpts...)),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	}
	conn, err := grpc.NewClient(fmt.Sprintf("%s:%d", v2Config.Host, v2Config.GRPCPort), opts...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func createV2ControlPlaneClient(v2Config V2Config) (v2.GRPCInferenceServiceClient, error) {
	conn, err := CreateV2GrpcConnection(v2Config)
	if err != nil {
		// TODO: this could fail in later iterations, so close earlier connections
		conn.Close()
		return nil, err
	}

	client := v2.NewGRPCInferenceServiceClient(conn)
	return client, nil
}

func GetV2ConfigWithDefaults(host string, port int) V2Config {
	return V2Config{
		Host:                         host,
		GRPCPort:                     port,
		GRPCRetryBackoff:             util.GRPCRetryBackoff,
		GRPRetryMaxCount:             util.GRPCRetryMaxCount,
		GRPCMaxMsgSizeBytes:          util.GRPCMaxMsgSizeBytes,
		GRPCModelServerLoadTimeout:   util.GRPCModelServerLoadTimeout,
		GRPCModelServerUnloadTimeout: util.GRPCModelServerUnloadTimeout,
		GRPCControlPlaneTimeout:      util.GRPCControlPlaneTimeout,
	}
}

func NewV2Client(v2Config V2Config, logger log.FieldLogger) *V2Client {
	logger.Infof("V2 (OIP) Inference Server %s:%d", v2Config.Host, v2Config.GRPCPort)

	grpcClient, err := createV2ControlPlaneClient(v2Config)
	if err != nil {
		return nil
	}

	return &V2Client{
		v2Config:   v2Config,
		grpcClient: grpcClient,
		logger:     logger.WithField("Source", "V2InferenceServerClientGrpc"),
	}

}

func (v *V2Client) LoadModel(name string) *interfaces.ControlPlaneErr {

	return v.loadModelGrpc(name)

}

func (v *V2Client) loadModelGrpc(name string) *interfaces.ControlPlaneErr {
	ctx, cancel := context.WithTimeout(context.Background(), v.v2Config.GRPCModelServerLoadTimeout)
	defer cancel()

	req := &v2.RepositoryModelLoadRequest{
		ModelName: name,
	}

	_, err := v.grpcClient.RepositoryModelLoad(ctx, req)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			errCode := e.Code()
			return &interfaces.ControlPlaneErr{
				Err:     err,
				ErrCode: int(errCode),
				IsGrpc:  true,
			}
		}
		return &interfaces.ControlPlaneErr{
			Err:     err,
			ErrCode: interfaces.V2CommunicationErrCode,
			IsGrpc:  true,
		}

	}
	return nil
}

func (v *V2Client) UnloadModel(name string) *interfaces.ControlPlaneErr {
	return v.unloadModelGrpc(name)
}

func (v *V2Client) unloadModelGrpc(name string) *interfaces.ControlPlaneErr {
	ctx, cancel := context.WithTimeout(context.Background(), v.v2Config.GRPCModelServerUnloadTimeout)
	defer cancel()

	req := &v2.RepositoryModelUnloadRequest{
		ModelName: name,
	}

	_, err := v.grpcClient.RepositoryModelUnload(ctx, req)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			errCode := e.Code()
			return &interfaces.ControlPlaneErr{
				Err:     err,
				ErrCode: int(errCode),
				IsGrpc:  true,
			}
		}
		return &interfaces.ControlPlaneErr{
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

	ready, err = v.liveGrpc()

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

func (v *V2Client) liveGrpc() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), v.v2Config.GRPCControlPlaneTimeout)
	defer cancel()

	req := &v2.ServerLiveRequest{}

	res, err := v.grpcClient.ServerLive(ctx, req)
	if err != nil {
		return false, err
	}
	return res.Live, nil
}

func (v *V2Client) GetModels() ([]interfaces.ServerModelInfo, error) {
	return v.getModelsGrpc()
}

func (v *V2Client) getModelsGrpc() ([]interfaces.ServerModelInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), v.v2Config.GRPCControlPlaneTimeout)
	defer cancel()

	var models []interfaces.ServerModelInfo
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
