package scheduler

import (
	"fmt"
	"math"
	"time"

	"google.golang.org/grpc/credentials/insecure"

	"github.com/seldonio/seldon-core-v2/components/tls/pkg/tls"

	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
)

type SchedulerClient struct {
	client.Client
	logger           logr.Logger
	conn             *grpc.ClientConn
	callOptions      []grpc.CallOption
	recorder         record.EventRecorder
	certificateStore *tls.CertificateStore
}

func NewSchedulerClient(logger logr.Logger, client client.Client, recorder record.EventRecorder) *SchedulerClient {
	opts := []grpc.CallOption{
		grpc.MaxCallSendMsgSize(math.MaxInt32),
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
	}

	return &SchedulerClient{
		Client:      client,
		logger:      logger.WithName("schedulerClient"),
		callOptions: opts,
		recorder:    recorder,
	}
}

func (s *SchedulerClient) ConnectToScheduler(host string, plainTxtPort int, tlsPort int) error {
	var err error
	protocol := tls.GetSecurityProtocolFromEnv(tls.EnvSecurityPrefixControlPlane)
	if protocol == tls.SecurityProtocolSSL {
		s.certificateStore, err = tls.NewCertificateStore(tls.Prefix(tls.EnvSecurityPrefixControlPlane))
		if err != nil {
			return err
		}
	}
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond)),
	}
	opts := []grpc.DialOption{}

	var port int
	if s.certificateStore != nil {
		port = tlsPort
		opts = append(opts, grpc.WithTransportCredentials(s.certificateStore.CreateClientTransportCredentials()))
		s.logger.Info("Running scheduler client in TLS mode", "port", port)
	} else {
		port = plainTxtPort
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		s.logger.Info("Running scheduler client in plain text mode", "port", port)
	}
	opts = append(opts, grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(retryOpts...)))
	opts = append(opts, grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOpts...)))

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), opts...)
	if err != nil {
		return err
	}
	s.conn = conn
	s.logger.Info("Connected to scheduler", "host", host, "port", port)
	return nil
}
