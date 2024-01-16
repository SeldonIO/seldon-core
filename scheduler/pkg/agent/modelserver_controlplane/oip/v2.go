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

type V2Client struct {
	grpcClient v2.GRPCInferenceServiceClient
	host       string
	grpcPort   int
	logger     log.FieldLogger
}

func CreateV2GrpcConnection(host string, plainTxtPort int) (*grpc.ClientConn, error) {
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(util.GrpcRetryBackoffMillisecs * time.Millisecond)),
		grpc_retry.WithMax(util.GrpcRetryMaxCount),
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(util.GrpcMaxMsgSizeBytes), grpc.MaxCallSendMsgSize(util.GrpcMaxMsgSizeBytes)),
		grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(retryOpts...)),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOpts...)),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	}
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, plainTxtPort), opts...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func createV2ControlPlaneClient(host string, port int) (v2.GRPCInferenceServiceClient, error) {
	conn, err := CreateV2GrpcConnection(host, port)
	if err != nil {
		// TODO: this could fail in later iterations, so close earlier connections
		conn.Close()
		return nil, err
	}

	client := v2.NewGRPCInferenceServiceClient(conn)
	return client, nil
}

func NewV2Client(host string, port int, logger log.FieldLogger) *V2Client {
	logger.Infof("V2 (OIP) Inference Server %s:%d", host, port)

	grpcClient, err := createV2ControlPlaneClient(host, port)
	if err != nil {
		return nil
	}

	return &V2Client{
		host:       host,
		grpcPort:   port,
		grpcClient: grpcClient,
		logger:     logger.WithField("Source", "V2InferenceServerClientGrpc"),
	}

}

func (v *V2Client) LoadModel(name string) *interfaces.ControlPlaneErr {

	return v.loadModelGrpc(name)

}

func (v *V2Client) loadModelGrpc(name string) *interfaces.ControlPlaneErr {
	ctx := context.Background()

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
	ctx := context.Background()

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
	ctx := context.Background()
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
