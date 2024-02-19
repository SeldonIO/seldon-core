/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package testing_utils

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/jarcoal/httpmock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v2 "github.com/seldonio/seldon-core/apis/go/v2/mlops/v2_dataplane"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
)

const (
	ModelNameMissing = "missingmodel"
)

type V2State struct {
	Models map[string]bool
	mu     sync.Mutex
}

func (s *V2State) LoadResponder(model string, status int) func(req *http.Request) (*http.Response, error) {
	return func(req *http.Request) (*http.Response, error) {
		s.SetModel(model, true)
		if status == 200 {
			return httpmock.NewStringResponse(status, ""), nil
		} else {
			return httpmock.NewStringResponse(status, ""), interfaces.ErrControlPlaneBadRequest
		}
	}
}

func (s *V2State) UnloadResponder(model string, status int) func(req *http.Request) (*http.Response, error) {
	return func(req *http.Request) (*http.Response, error) {
		s.SetModel(model, false)
		if status == 200 {
			return httpmock.NewStringResponse(status, ""), nil
		} else {
			return httpmock.NewStringResponse(status, ""), interfaces.ErrControlPlaneBadRequest
		}
	}
}

func (s *V2State) SetModel(modelId string, val bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Models[modelId] = val
}

func (s *V2State) IsModelLoaded(modelId string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	val, loaded := s.Models[modelId]
	if loaded {
		return val
	}
	return false
}

type MockGRPCMLServer struct {
	listener    net.Listener
	server      *grpc.Server
	models      []interfaces.ServerModelInfo
	isReady     bool
	LoadSleep   time.Duration
	UnloadSleep time.Duration
	v2.UnimplementedGRPCInferenceServiceServer
}

func (m *MockGRPCMLServer) Setup(port uint) error {
	var err error
	m.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	opts := []grpc.ServerOption{}
	m.server = grpc.NewServer(opts...)
	v2.RegisterGRPCInferenceServiceServer(m.server, m)
	return nil
}

func (m *MockGRPCMLServer) Start() error {
	m.isReady = true
	return m.server.Serve(m.listener)
}

func (m *MockGRPCMLServer) Stop() {
	m.isReady = false
	m.server.Stop()
}

func (m *MockGRPCMLServer) ModelInfer(ctx context.Context, r *v2.ModelInferRequest) (*v2.ModelInferResponse, error) {
	return &v2.ModelInferResponse{ModelName: r.ModelName, ModelVersion: r.ModelVersion}, nil
}

func (m *MockGRPCMLServer) ModelMetadata(ctx context.Context, r *v2.ModelMetadataRequest) (*v2.ModelMetadataResponse, error) {
	return &v2.ModelMetadataResponse{Name: r.Name, Versions: []string{r.Version}}, nil
}

func (m *MockGRPCMLServer) ModelReady(ctx context.Context, r *v2.ModelReadyRequest) (*v2.ModelReadyResponse, error) {
	return &v2.ModelReadyResponse{Ready: m.isReady}, nil
}

func (m *MockGRPCMLServer) ServerReady(ctx context.Context, r *v2.ServerReadyRequest) (*v2.ServerReadyResponse, error) {
	return &v2.ServerReadyResponse{Ready: true}, nil
}

func (m *MockGRPCMLServer) ServerLive(ctx context.Context, r *v2.ServerLiveRequest) (*v2.ServerLiveResponse, error) {
	return &v2.ServerLiveResponse{Live: true}, nil
}

func (m *MockGRPCMLServer) RepositoryModelLoad(ctx context.Context, r *v2.RepositoryModelLoadRequest) (*v2.RepositoryModelLoadResponse, error) {
	time.Sleep(m.LoadSleep)
	return &v2.RepositoryModelLoadResponse{}, nil
}

func (m *MockGRPCMLServer) RepositoryModelUnload(ctx context.Context, r *v2.RepositoryModelUnloadRequest) (*v2.RepositoryModelUnloadResponse, error) {
	time.Sleep(m.UnloadSleep)
	if r.ModelName == ModelNameMissing {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Model %s not found", r.ModelName))
	}
	return &v2.RepositoryModelUnloadResponse{}, nil
}

func (m *MockGRPCMLServer) RepositoryIndex(ctx context.Context, r *v2.RepositoryIndexRequest) (*v2.RepositoryIndexResponse, error) {
	ret := make([]*v2.RepositoryIndexResponse_ModelIndex, len(m.models))
	for idx, model := range m.models {
		ret[idx] = &v2.RepositoryIndexResponse_ModelIndex{Name: model.Name, State: string(model.State)}
	}
	return &v2.RepositoryIndexResponse{Models: ret}, nil
}

func (m *MockGRPCMLServer) SetModels(models []interfaces.ServerModelInfo) {
	m.models = models
}
