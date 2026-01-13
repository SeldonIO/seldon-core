/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package db_service

import (
	"context"
	"fmt"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler/db"
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler/dbsvc"
	seldontls "github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"
	grpcsvc "github.com/seldonio/seldon-core/scheduler/v2/pkg/grpc"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type databaseService struct {
	dbsvc.UnimplementedDatabaseServiceServer
	log     *logrus.Entry
	models  store.StorageReader[*db.Model]
	servers store.StorageReader[*db.Server]
}

func newDatabaseService(
	logger *logrus.Entry,
	models store.StorageReader[*db.Model],
	servers store.StorageReader[*db.Server],
) *databaseService {
	return &databaseService{
		log:     logger,
		models:  models,
		servers: servers,
	}
}

func NewDatabaseService(
	logger *logrus.Entry,
	models store.StorageReader[*db.Model],
	servers store.StorageReader[*db.Server],
	tlsOptions seldontls.TLSOptions,
) *grpcsvc.Server {
	impl := newDatabaseService(logger, models, servers)
	return grpcsvc.NewServer(logger, tlsOptions, func(grpcServer *grpc.Server) {
		dbsvc.RegisterDatabaseServiceServer(grpcServer, impl)
	})
}

func (d *databaseService) DumpDatabase(ctx context.Context, req *dbsvc.DatabaseDumpRequest) (*dbsvc.DatabaseDumpResponse, error) {
	log := d.log.WithField("func", "DumpDatabase")
	log.Info("DB dump requested")

	resp := &dbsvc.DatabaseDumpResponse{}
	var err error
	switch req.Action {
	case dbsvc.DatabaseDumpRequest_DUMP_MODELS:
		resp.Model, err = d.models.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("list models: %w", err)
		}
	case dbsvc.DatabaseDumpRequest_DUMP_SERVERS:
		resp.Server, err = d.servers.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("list servers: %w", err)
		}
	case dbsvc.DatabaseDumpRequest_DUMP_ALL:
		resp.Server, err = d.servers.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("list servers: %w", err)
		}
		resp.Model, err = d.models.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("list models: %w", err)
		}
	default:
		return nil, status.Errorf(codes.InvalidArgument, "invalid dump request action: %d", req.Action)
	}

	return resp, nil
}
