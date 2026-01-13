/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package cli

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"math"
	"os"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler/dbsvc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type DbClient struct {
	schedulerHost string
	authority     string
	callOptions   []grpc.CallOption
	config        *SeldonCLIConfig
	verbose       bool
}

func NewDBClient(schedulerHost string, schedulerHostIsSet bool, authority string, verbose bool) (*DbClient, error) {
	opts := []grpc.CallOption{
		grpc.MaxCallSendMsgSize(math.MaxInt32),
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
	}
	config, err := LoadSeldonCLIConfig()
	if err != nil {
		return nil, err
	}

	// Overwrite host if set in config
	if !schedulerHostIsSet && config.Controlplane != nil && config.Controlplane.SchedulerHost != "" {
		schedulerHost = config.Controlplane.SchedulerHost
	}
	return &DbClient{
		schedulerHost: schedulerHost,
		authority:     authority,
		callOptions:   opts,
		config:        config,
		verbose:       verbose,
	}, nil
}

func (d *DbClient) loadKeyPair() (credentials.TransportCredentials, error) {
	certificate, err := tls.LoadX509KeyPair(d.config.Controlplane.CrtPath, d.config.Controlplane.KeyPath)
	if err != nil {
		return nil, err
	}

	ca, err := os.ReadFile(d.config.Controlplane.CaPath)
	if err != nil {
		return nil, err
	}

	capool := x509.NewCertPool()
	if !capool.AppendCertsFromPEM(ca) {
		return nil, fmt.Errorf("Failed to load ca crt from %s", d.config.Controlplane.CaPath)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		RootCAs:      capool,
	}

	return credentials.NewTLS(tlsConfig), nil
}

func (d *DbClient) newConnection() (*grpc.ClientConn, error) {
	var creds credentials.TransportCredentials
	if d.config.Controlplane == nil || d.config.Controlplane.KeyPath == "" {
		creds = insecure.NewCredentials()
	} else {
		tlsCreds, err := d.loadKeyPair()
		if err != nil {
			return nil, err
		}
		creds = tlsCreds
	}

	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond)),
	}
	unaryInterceptor := grpc_retry.UnaryClientInterceptor(retryOpts...)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithUnaryInterceptor(unaryInterceptor),
		grpc.WithAuthority(d.authority),
	}

	conn, err := grpc.NewClient(d.schedulerHost, opts...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

type DBRecordType string

const (
	DBRecordTypeAll     DBRecordType = "all"     // 1
	DBRecordTypeServers              = "servers" // 2
	DBRecordTypeModels               = "models"  // 3
)

type DBDumpOutputFormat string

const (
	DBDumpOutputFormatJson  DBDumpOutputFormat = "json"
	DBDumpOutputFormatProto DBDumpOutputFormat = "proto"
)

type DumpDBRequest struct {
	Action       DBRecordType
	OutputFile   string
	OutputFormat DBDumpOutputFormat
}

func (d *DbClient) DumpDatabase(r DumpDBRequest) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req := &dbsvc.DatabaseDumpRequest{}
	if d.verbose {
		printProto(req)
	}
	conn, err := d.newConnection()
	if err != nil {
		return err
	}
	grpcClient := dbsvc.NewDatabaseServiceClient(conn)

	if r.OutputFile == "" {
		return fmt.Errorf("no output file specified")
	}

	switch r.Action {
	case DBRecordTypeAll:
		req.Action = dbsvc.DatabaseDumpRequest_DUMP_ALL
	case DBRecordTypeServers:
		req.Action = dbsvc.DatabaseDumpRequest_DUMP_SERVERS
	case DBRecordTypeModels:
		req.Action = dbsvc.DatabaseDumpRequest_DUMP_MODELS
	default:
		return fmt.Errorf("unknown action %s", r.Action)
	}

	resp, err := grpcClient.DumpDatabase(ctx, req)
	if err != nil {
		return fmt.Errorf("gRPC dump request failed: %w", err)
	}

	var data []byte
	switch r.OutputFormat {
	case DBDumpOutputFormatProto:
		bb, err := proto.Marshal(resp)
		if err != nil {
			return fmt.Errorf("marshaling response failed: %w", err)
		}
		data = bb
	case DBDumpOutputFormatJson:
		jsonBytes, err := protojson.Marshal(resp)
		if err != nil {
			return fmt.Errorf("marshaling response to JSON failed: %w", err)
		}
		data = jsonBytes
	default:
		return fmt.Errorf("unsupported output format: %s", r.OutputFormat)
	}

	return writeToFileIfEmpty(r.OutputFile, data)
}

func writeToFileIfEmpty(filePath string, data []byte) error {
	if fileInfo, err := os.Stat(filePath); err == nil {
		if fileInfo.Size() > 0 {
			return fmt.Errorf("output file %s already exists and is not empty", filePath)
		}
	}
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write to output file %s: %w", filePath, err)
	}
	return nil
}
