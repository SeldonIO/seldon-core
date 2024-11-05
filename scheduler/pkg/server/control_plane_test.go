/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package server

import (
	"context"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/internal/testing_utils"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/synchroniser"
)

func TestStartServerStream(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name   string
		server *SchedulerServer
		err    bool
	}

	tests := []test{
		{
			name: "ok",
			server: &SchedulerServer{
				modelStore: store.NewMemoryStore(log.New(), store.NewLocalSchedulerStore(), nil),
				logger:     log.New(),
				timeout:    10 * time.Millisecond,
			},
		},
		{
			name: "timeout",
			server: &SchedulerServer{
				modelStore: store.NewMemoryStore(log.New(), store.NewLocalSchedulerStore(), nil),
				logger:     log.New(),
				timeout:    1 * time.Millisecond,
			},
			err: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stream := newStubControlPlaneServer(1, 5*time.Millisecond)
			err := test.server.sendStartServerStreamMarker(stream)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())

				var msr *pb.ControlPlaneResponse
				select {
				case next := <-stream.msgs:
					msr = next
				default:
					t.Fail()
				}

				g.Expect(msr).ToNot(BeNil())
				g.Expect(msr.Event).To(Equal(pb.ControlPlaneResponse_SEND_SERVERS))
			}

			err = test.server.sendResourcesMarker(stream)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())

				var msr *pb.ControlPlaneResponse
				select {
				case next := <-stream.msgs:
					msr = next
				default:
					t.Fail()
				}

				g.Expect(msr).ToNot(BeNil())
				g.Expect(msr.Event).To(Equal(pb.ControlPlaneResponse_SEND_RESOURCES))
			}
		})
	}
}

func TestSubscribeControlPlane(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	g := NewGomegaWithT(t)

	type test struct {
		name string
	}
	tests := []test{
		{
			name: "simple",
		},
	}

	getStream := func(context context.Context, port int) (*grpc.ClientConn, pb.Scheduler_SubscribeControlPlaneClient) {
		conn, _ := grpc.NewClient(fmt.Sprintf(":%d", port), grpc.WithTransportCredentials(insecure.NewCredentials()))
		grpcClient := pb.NewSchedulerClient(conn)
		client, _ := grpcClient.SubscribeControlPlane(
			context,
			&pb.ControlPlaneSubscriptionRequest{SubscriberName: "dummy"},
		)
		return conn, client
	}

	createTestScheduler := func() *SchedulerServer {
		logger := log.New()
		logger.SetLevel(log.WarnLevel)

		eventHub, err := coordinator.NewEventHub(logger)
		g.Expect(err).To(BeNil())

		sync := synchroniser.NewSimpleSynchroniser(time.Duration(10 * time.Millisecond))

		s := NewSchedulerServer(logger, nil, nil, nil, nil, eventHub, sync)
		sync.Signals(1)

		return s
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := createTestScheduler()
			port, err := testing_utils.GetFreePortForTest()
			if err != nil {
				t.Fatal(err)
			}
			err = server.startServer(uint(port), false)
			if err != nil {
				t.Fatal(err)
			}
			time.Sleep(100 * time.Millisecond)

			conn, client := getStream(context.Background(), port)

			msg, _ := client.Recv()
			g.Expect(msg.GetEvent()).To(Equal(pb.ControlPlaneResponse_SEND_SERVERS))

			msg, _ = client.Recv()
			g.Expect(msg.Event).To(Equal(pb.ControlPlaneResponse_SEND_RESOURCES))

			conn.Close()
			server.StopSendControlPlaneEvents()
		})
	}
}
