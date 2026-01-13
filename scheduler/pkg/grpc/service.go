package grpc

import (
	"fmt"
	"net"

	seldontls "github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// TODO we should migrate all other gRPC services to use this pkg

// Server manages the gRPC server lifecycle for any gRPC service
type Server struct {
	log          *logrus.Entry
	grpcServer   *grpc.Server
	tlsOptions   seldontls.TLSOptions
	registerFunc func(*grpc.Server)
}

func NewServer(
	logger *logrus.Entry,
	tlsOptions seldontls.TLSOptions,
	registerFunc func(*grpc.Server),
) *Server {
	return &Server{
		log:          logger,
		tlsOptions:   tlsOptions,
		registerFunc: registerFunc,
	}
}

func (s *Server) start(port uint, secure bool) error {
	logger := s.log.WithField("func", "start")
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
	s.registerFunc(grpcServer)

	go func() {
		s.log.Infof("gRPC service server running on port %d mtls:%v", port, secure)
		if err := grpcServer.Serve(lis); err != nil {
			logger.WithError(err).Fatalf("gRPC server failed on port %d mtls:%v", port, secure)
		}
	}()

	return nil
}

func (s *Server) Start(allowPlainTxt bool, port uint) error {
	logger := s.log.WithField("func", "Start")

	if !allowPlainTxt && s.tlsOptions.Cert == nil {
		return fmt.Errorf("one of plain txt or mTLS needs to be defined. But have plain text [%v] and no TLS", allowPlainTxt)
	}

	if allowPlainTxt {
		err := s.start(port, false)
		if err != nil {
			return err
		}
		return nil
	}
	logger.Info("Not starting plain text gRPC server")

	if s.tlsOptions.Cert != nil {
		err := s.start(port, true)
		if err != nil {
			return err
		}
		return nil
	}

	logger.Info("Not starting mTLS grpc server")
	return nil
}

func (s *Server) Stop() {
	if s.grpcServer != nil {
		s.log.Warn("Gracefully stopping gRPC server")
		s.grpcServer.GracefulStop()
	}
}
