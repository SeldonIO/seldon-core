package grpc

import (
	"net"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	seldontls "github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func TestServer_Start_Table(t *testing.T) {
	type testCase struct {
		name          string
		allowPlainTxt bool
		tlsOptions    seldontls.TLSOptions
		port          func(g Gomega) uint
		expectErr     bool
		expectStarted bool
		expectRegCall bool
		doStop        bool
	}

	tests := []testCase{
		{
			name:          "plain text allowed, no TLS: starts, registers, and stops",
			allowPlainTxt: true,
			tlsOptions:    seldontls.TLSOptions{Cert: nil},
			port: func(g Gomega) uint {
				// Use ephemeral port :0
				return 0
			},
			expectErr:     false,
			expectStarted: true,
			expectRegCall: true,
			doStop:        true,
		},
		{
			name:          "plain text disallowed and no TLS cert: returns error",
			allowPlainTxt: false,
			tlsOptions:    seldontls.TLSOptions{Cert: nil},
			port: func(g Gomega) uint {
				return 0
			},
			expectErr:     true,
			expectStarted: false,
			expectRegCall: false,
			doStop:        true, // should be safe no-op
		},
		{
			name:          "port already in use: returns listen error (does not start/register)",
			allowPlainTxt: true,
			tlsOptions:    seldontls.TLSOptions{Cert: nil},
			port: func(g Gomega) uint {
				lis, err := net.Listen("tcp", ":0")
				g.Expect(err).NotTo(HaveOccurred())
				t.Cleanup(func() { _ = lis.Close() })

				addr, ok := lis.Addr().(*net.TCPAddr)
				g.Expect(ok).To(BeTrue())
				g.Expect(addr.Port).To(BeNumerically(">", 0))
				return uint(addr.Port)
			},
			expectErr:     true,
			expectStarted: false,
			expectRegCall: false,
			doStop:        true, // should be safe no-op
		},
	}

	for _, tt := range tests {
		tt := tt // capture
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			logger := logrus.New()
			entry := logrus.NewEntry(logger)

			registered := false
			registerFunc := func(s *grpc.Server) {
				registered = true
				g.Expect(s).NotTo(BeNil())
			}

			srv := NewServer(entry, tt.tlsOptions, registerFunc)

			err := srv.Start(tt.allowPlainTxt, tt.port(g))

			if tt.expectErr {
				g.Expect(err).To(HaveOccurred())
				g.Expect(srv.grpcServer).To(BeNil())
				g.Expect(registered).To(Equal(tt.expectRegCall))
				if tt.doStop {
					g.Expect(func() { srv.Stop() }).NotTo(Panic())
				}
				return
			}

			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(srv.grpcServer).NotTo(BeNil())
			g.Expect(registered).To(Equal(tt.expectRegCall))

			// Give the goroutine a moment to enter Serve() before stopping.
			time.Sleep(10 * time.Millisecond)

			if tt.doStop {
				g.Expect(func() { srv.Stop() }).NotTo(Panic())
			}
		})
	}
}

func TestServer_Stop_NoStart_DoesNotPanic(t *testing.T) {
	g := NewWithT(t)

	logger := logrus.NewEntry(logrus.New())
	srv := NewServer(logger, seldontls.TLSOptions{Cert: nil}, func(*grpc.Server) {})

	g.Expect(func() { srv.Stop() }).NotTo(Panic())
}
