package db_service

import (
	"context"
	"fmt"
	"net"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler/db"
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler/dbdump"
	seldontls "github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Service struct {
	dbdump.UnimplementedDatabaseServiceServer
	log        *logrus.Entry
	grpcServer *grpc.Server
	tlsOptions seldontls.TLSOptions
	api        struct {
		models  store.StorageReader[*db.Model]
		servers store.StorageReader[*db.Server]
	}
}

func NewService(
	logger *logrus.Entry,
	models store.StorageReader[*db.Model],
	servers store.StorageReader[*db.Server],
	tlsOptions seldontls.TLSOptions,
) *Service {
	return &Service{
		log:        logger,
		tlsOptions: tlsOptions,
		api: struct {
			models  store.StorageReader[*db.Model]
			servers store.StorageReader[*db.Server]
		}{
			models:  models,
			servers: servers,
		},
	}
}

func (s *Service) startServer(port uint, secure bool) error {
	logger := s.log.WithField("func", "startServer")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	opts := []grpc.ServerOption{}
	if secure {
		opts = append(opts, grpc.Creds(s.tlsOptions.Cert.CreateServerTransportCredentials()))
	}

	grpcServer := grpc.NewServer(opts...)
	s.grpcServer = grpcServer
	dbdump.RegisterDatabaseServiceServer(grpcServer, s)

	s.log.Infof("Database service server running on port %d mtls:%v", port, secure)
	go func() {
		err := grpcServer.Serve(lis)
		logger.WithError(err).Fatalf("Database service server failed on port %d mtls:%v", port, secure)
	}()

	return nil
}

func (s *Service) StartGrpcServer(allowPlainTxt bool, port uint) error {
	logger := s.log.WithField("func", "StartGrpcServers")

	if !allowPlainTxt && s.tlsOptions.Cert == nil {
		return fmt.Errorf("one of plain txt or mTLS needs to be defined. But have plain text [%v] and no TLS", allowPlainTxt)
	}

	if allowPlainTxt {
		err := s.startServer(port, false)
		if err != nil {
			return err
		}
		return nil
	}
	logger.Info("Not starting database service plain text server")

	if s.tlsOptions.Cert != nil {
		err := s.startServer(port, true)
		if err != nil {
			return err
		}
		return nil
	}

	logger.Info("Not starting database service mTLS server")
	return nil
}

func (s *Service) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
		s.log.Info("Database service closing gRPC server")
	}
}

func (s *Service) DumpDatabase(ctx context.Context, req *dbdump.DatabaseDumpRequest) (*dbdump.DatabaseDumpResponse, error) {
	log := s.log.WithField("func", "DumpDatabase")
	log.Info("DB dump requested")

	resp := &dbdump.DatabaseDumpResponse{}
	var err error
	switch req.Action {
	case dbdump.DatabaseDumpRequest_DUMP_MODELS:
		resp.Model, err = s.api.models.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("list models: %w", err)
		}
	case dbdump.DatabaseDumpRequest_DUMP_SERVERS:
		resp.Server, err = s.api.servers.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("list servers: %w", err)
		}
	case dbdump.DatabaseDumpRequest_DUMP_ALL:
		resp.Server, err = s.api.servers.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("list servers: %w", err)
		}
		resp.Model, err = s.api.models.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("list models: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported dump request action %v", req.Action)
	}

	return resp, nil
}
