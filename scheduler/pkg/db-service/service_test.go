package db_service

import (
	"context"
	"errors"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler/db"
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler/dbsvc"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/mock"
)

func TestDumpDatabase(t *testing.T) {
	type mockSetup func(
		ctx context.Context,
		models *mock.MockStorageReader[*db.Model],
		servers *mock.MockStorageReader[*db.Server],
	)

	tests := []struct {
		name           string
		action         dbsvc.DatabaseDumpRequest_DatabaseDumpRecords
		setupMocks     mockSetup
		expectErr      bool
		expectGRPCCode codes.Code
		expectModels   []string
		expectServers  []string
	}{
		{
			name:   "dump models success",
			action: dbsvc.DatabaseDumpRequest_DUMP_MODELS,
			setupMocks: func(ctx context.Context, m *mock.MockStorageReader[*db.Model], s *mock.MockStorageReader[*db.Server]) {
				m.EXPECT().List(ctx).Return([]*db.Model{
					{Name: "model1"},
					{Name: "model2"},
				}, nil)
			},
			expectModels: []string{"model1", "model2"},
		},
		{
			name:   "dump servers success",
			action: dbsvc.DatabaseDumpRequest_DUMP_SERVERS,
			setupMocks: func(ctx context.Context, m *mock.MockStorageReader[*db.Model], s *mock.MockStorageReader[*db.Server]) {
				s.EXPECT().List(ctx).Return([]*db.Server{
					{Name: "server1"},
					{Name: "server2"},
				}, nil)
			},
			expectServers: []string{"server1", "server2"},
		},
		{
			name:   "dump all success",
			action: dbsvc.DatabaseDumpRequest_DUMP_ALL,
			setupMocks: func(ctx context.Context, m *mock.MockStorageReader[*db.Model], s *mock.MockStorageReader[*db.Server]) {
				s.EXPECT().List(ctx).Return([]*db.Server{{Name: "server1"}}, nil)
				m.EXPECT().List(ctx).Return([]*db.Model{{Name: "model1"}}, nil)
			},
			expectModels:  []string{"model1"},
			expectServers: []string{"server1"},
		},
		{
			name:   "dump models list error",
			action: dbsvc.DatabaseDumpRequest_DUMP_MODELS,
			setupMocks: func(ctx context.Context, m *mock.MockStorageReader[*db.Model], s *mock.MockStorageReader[*db.Server]) {
				m.EXPECT().List(ctx).Return(nil, errors.New("database error"))
			},
			expectErr: true,
		},
		{
			name:   "dump servers list error",
			action: dbsvc.DatabaseDumpRequest_DUMP_SERVERS,
			setupMocks: func(ctx context.Context, m *mock.MockStorageReader[*db.Model], s *mock.MockStorageReader[*db.Server]) {
				s.EXPECT().List(ctx).Return(nil, errors.New("database error"))
			},
			expectErr: true,
		},
		{
			name:   "dump all servers list error",
			action: dbsvc.DatabaseDumpRequest_DUMP_ALL,
			setupMocks: func(ctx context.Context, m *mock.MockStorageReader[*db.Model], s *mock.MockStorageReader[*db.Server]) {
				s.EXPECT().List(ctx).Return(nil, errors.New("database error"))
			},
			expectErr: true,
		},
		{
			name:   "dump all models list error",
			action: dbsvc.DatabaseDumpRequest_DUMP_ALL,
			setupMocks: func(ctx context.Context, m *mock.MockStorageReader[*db.Model], s *mock.MockStorageReader[*db.Server]) {
				s.EXPECT().List(ctx).Return([]*db.Server{{Name: "server1"}}, nil)
				m.EXPECT().List(ctx).Return(nil, errors.New("database error"))
			},
			expectErr: true,
		},
		{
			name:           "invalid action",
			action:         999,
			expectErr:      true,
			expectGRPCCode: codes.InvalidArgument,
		},
		{
			name:   "empty results",
			action: dbsvc.DatabaseDumpRequest_DUMP_ALL,
			setupMocks: func(ctx context.Context, m *mock.MockStorageReader[*db.Model], s *mock.MockStorageReader[*db.Server]) {
				s.EXPECT().List(ctx).Return([]*db.Server{}, nil)
				m.EXPECT().List(ctx).Return([]*db.Model{}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			ctrl := gomock.NewController(t)
			ctx := context.Background()
			mockModels := mock.NewMockStorageReader[*db.Model](ctrl)
			mockServers := mock.NewMockStorageReader[*db.Server](ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(ctx, mockModels, mockServers)
			}

			svc := newDatabaseService(
				logrus.NewEntry(logrus.New()),
				mockModels,
				mockServers,
			)

			resp, err := svc.DumpDatabase(ctx, &dbsvc.DatabaseDumpRequest{
				Action: tt.action,
			})

			if tt.expectErr {
				g.Expect(err).To(HaveOccurred())
				g.Expect(resp).To(BeNil())

				if tt.expectGRPCCode != codes.OK {
					st, ok := status.FromError(err)
					g.Expect(ok).To(BeTrue())
					g.Expect(st.Code()).To(Equal(tt.expectGRPCCode))
				}
				return
			}

			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(resp).NotTo(BeNil())

			g.Expect(resp.Model).To(HaveLen(len(tt.expectModels)))
			for i, name := range tt.expectModels {
				g.Expect(resp.Model[i].Name).To(Equal(name))
			}

			g.Expect(resp.Server).To(HaveLen(len(tt.expectServers)))
			for i, name := range tt.expectServers {
				g.Expect(resp.Server[i].Name).To(Equal(name))
			}
		})
	}
}
