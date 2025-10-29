package heartbeat

import (
	"context"
	"errors"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/heartbeat/mock"
)

func TestManager_CheckHeartbeats_Success(t *testing.T) {
	log := logrus.New()
	manager := NewManager(log)

	ctrl := gomock.NewController(t)
	mockSvc := mock.NewMockService(ctrl)
	requestChan := make(chan struct{}, 1)
	heartbeatChan := make(chan error, 1)

	mockSvc.EXPECT().Name().Return("test-service")
	mockSvc.EXPECT().RequestHeartBeat().Return(requestChan)
	mockSvc.EXPECT().Heartbeat().Return(heartbeatChan)

	manager.Register(mockSvc)

	go func() {
		<-requestChan
		heartbeatChan <- nil
	}()

	ctx := context.Background()
	err := manager.CheckHeartbeats(ctx)

	assert.NoError(t, err)
}

func TestManager_CheckHeartbeats_MultipleServices_Success(t *testing.T) {
	log := logrus.New()
	manager := NewManager(log)

	ctrl := gomock.NewController(t)
	mockSvc1 := mock.NewMockService(ctrl)
	mockSvc2 := mock.NewMockService(ctrl)

	requestChan1 := make(chan struct{}, 1)
	heartbeatChan1 := make(chan error, 1)
	requestChan2 := make(chan struct{}, 1)
	heartbeatChan2 := make(chan error, 1)

	mockSvc1.EXPECT().Name().Return("service-1")
	mockSvc1.EXPECT().RequestHeartBeat().Return(requestChan1)
	mockSvc1.EXPECT().Heartbeat().Return(heartbeatChan1)

	mockSvc2.EXPECT().Name().Return("service-2")
	mockSvc2.EXPECT().RequestHeartBeat().Return(requestChan2)
	mockSvc2.EXPECT().Heartbeat().Return(heartbeatChan2)

	manager.Register(mockSvc1, mockSvc2)

	go func() {
		<-requestChan1
		heartbeatChan1 <- nil
	}()
	go func() {
		<-requestChan2
		heartbeatChan2 <- nil
	}()

	ctx := context.Background()
	err := manager.CheckHeartbeats(ctx)

	assert.NoError(t, err)
}

func TestManager_CheckHeartbeats_ServiceError(t *testing.T) {
	log := logrus.New()
	manager := NewManager(log)

	ctrl := gomock.NewController(t)
	mockSvc := mock.NewMockService(ctrl)
	requestChan := make(chan struct{}, 1)
	heartbeatChan := make(chan error, 1)

	mockSvc.EXPECT().Name().Return("test-service").MinTimes(1)
	mockSvc.EXPECT().RequestHeartBeat().Return(requestChan)
	mockSvc.EXPECT().Heartbeat().Return(heartbeatChan)

	manager.Register(mockSvc)

	expectedErr := errors.New("service unavailable")

	go func() {
		<-requestChan
		heartbeatChan <- expectedErr
	}()

	ctx := context.Background()
	err := manager.CheckHeartbeats(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "heartbeat failed for test-service")
	assert.Contains(t, err.Error(), "service unavailable")
}

func TestManager_CheckHeartbeats_ChannelClosed(t *testing.T) {
	log := logrus.New()
	manager := NewManager(log)

	ctrl := gomock.NewController(t)
	mockSvc := mock.NewMockService(ctrl)
	requestChan := make(chan struct{}, 1)
	heartbeatChan := make(chan error)

	mockSvc.EXPECT().Name().Return("test-service").MinTimes(1)
	mockSvc.EXPECT().RequestHeartBeat().Return(requestChan)
	mockSvc.EXPECT().Heartbeat().Return(heartbeatChan)

	manager.Register(mockSvc)

	go func() {
		<-requestChan
		close(heartbeatChan)
	}()

	ctx := context.Background()
	err := manager.CheckHeartbeats(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "heartbeat channel closed for test-service")
}

func TestManager_CheckHeartbeats_ContextCancelled(t *testing.T) {
	log := logrus.New()
	manager := NewManager(log)

	ctrl := gomock.NewController(t)
	mockSvc := mock.NewMockService(ctrl)
	requestChan := make(chan struct{}, 1)
	heartbeatChan := make(chan error)

	mockSvc.EXPECT().RequestHeartBeat().Return(requestChan)
	mockSvc.EXPECT().Heartbeat().Return(heartbeatChan)

	manager.Register(mockSvc)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-requestChan
		cancel()
	}()

	err := manager.CheckHeartbeats(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestManager_CheckHeartbeats_OneServiceFails(t *testing.T) {
	log := logrus.New()
	manager := NewManager(log)

	ctrl := gomock.NewController(t)
	mockSvc1 := mock.NewMockService(ctrl)
	mockSvc2 := mock.NewMockService(ctrl)

	requestChan1 := make(chan struct{}, 1)
	heartbeatChan1 := make(chan error, 1)
	requestChan2 := make(chan struct{}, 1)
	heartbeatChan2 := make(chan error, 1)

	mockSvc1.EXPECT().Name().Return("service-1").AnyTimes()
	mockSvc1.EXPECT().RequestHeartBeat().Return(requestChan1)
	mockSvc1.EXPECT().Heartbeat().Return(heartbeatChan1).AnyTimes()

	mockSvc2.EXPECT().Name().Return("service-2").AnyTimes()
	mockSvc2.EXPECT().RequestHeartBeat().Return(requestChan2)
	mockSvc2.EXPECT().Heartbeat().Return(heartbeatChan2)

	manager.Register(mockSvc1, mockSvc2)

	expectedErr := errors.New("connection timeout")

	go func() {
		<-requestChan1
		heartbeatChan1 <- nil
	}()
	go func() {
		<-requestChan2
		heartbeatChan2 <- expectedErr
	}()

	ctx := context.Background()
	err := manager.CheckHeartbeats(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "heartbeat failed for service-2")
	assert.Contains(t, err.Error(), "connection timeout")
}

func TestManager_CheckHeartbeats_NoServices(t *testing.T) {
	log := logrus.New()
	manager := NewManager(log)

	ctx := context.Background()
	err := manager.CheckHeartbeats(ctx)

	assert.NoError(t, err)
}

func TestManager_CheckHeartbeats_MultipleErrors(t *testing.T) {
	log := logrus.New()
	manager := NewManager(log)

	ctrl := gomock.NewController(t)
	mockSvc1 := mock.NewMockService(ctrl)
	mockSvc2 := mock.NewMockService(ctrl)

	requestChan1 := make(chan struct{}, 1)
	heartbeatChan1 := make(chan error, 1)
	requestChan2 := make(chan struct{}, 1)
	heartbeatChan2 := make(chan error, 1)

	mockSvc1.EXPECT().Name().Return("service-1").AnyTimes()
	mockSvc1.EXPECT().RequestHeartBeat().Return(requestChan1)
	mockSvc1.EXPECT().Heartbeat().Return(heartbeatChan1).AnyTimes()

	mockSvc2.EXPECT().Name().Return("service-2").AnyTimes()
	mockSvc2.EXPECT().RequestHeartBeat().Return(requestChan2)
	mockSvc2.EXPECT().Heartbeat().Return(heartbeatChan2).AnyTimes()

	manager.Register(mockSvc1, mockSvc2)

	err1 := errors.New("error 1")
	err2 := errors.New("error 2")

	go func() {
		<-requestChan1
		heartbeatChan1 <- err1
	}()
	go func() {
		<-requestChan2
		heartbeatChan2 <- err2
	}()

	ctx := context.Background()
	err := manager.CheckHeartbeats(ctx)

	require.Error(t, err)
	// errgroup returns the first error it encounters
	assert.Contains(t, err.Error(), "failed waiting for heartbeats")
}
