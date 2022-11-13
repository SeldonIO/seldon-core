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

package scheduler

import (
	"fmt"
	"math"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
		s.certificateStore, err = tls.NewCertificateStore(tls.Prefix(tls.EnvSecurityPrefixControlPlaneClient),
			tls.ValidationPrefix(tls.EnvSecurityPrefixControlPlaneServer))
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

func (s *SchedulerClient) checkErrorRetryable(resource string, resourceName string, err error) bool {
	if err != nil {
		if st, ok := status.FromError(err); ok {
			s.logger.Info("Got grpc status code", "err", err.Error(), "code", st.Code(), "resource", resource, "resourceName", resourceName)
			switch st.Code() {
			case codes.FailedPrecondition,
				codes.Unimplemented:
				s.logger.Info("Non retryable error", "code", st.Code(), "resource", resource, "resourceName", resourceName)
				return false
			default:
				s.logger.Info("retryable error", "code", st.Code(), "resource", resource, "resourceName", resourceName)
				return true
			}
		} else {
			s.logger.Info("Got non grpc error", "error", err.Error(), "resource", resource, "resourceName", resourceName)
			return true
		}
	} else {
		return false
	}

}
