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

package cli

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"text/tabwriter"
	"time"

	"github.com/ghodss/yaml"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
)

const subscriberName = "seldon CLI"

type SchedulerClient struct {
	schedulerHost string
	authority     string
	callOptions   []grpc.CallOption
	config        *SeldonCLIConfig
	verbose       bool
}

func NewSchedulerClient(schedulerHost string, schedulerHostIsSet bool, authority string, verbose bool) (*SchedulerClient, error) {

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
	return &SchedulerClient{
		schedulerHost: schedulerHost,
		authority:     authority,
		callOptions:   opts,
		config:        config,
		verbose:       verbose,
	}, nil
}

func (sc *SchedulerClient) loadKeyPair() (credentials.TransportCredentials, error) {
	certificate, err := tls.LoadX509KeyPair(sc.config.Controlplane.CrtPath, sc.config.Controlplane.KeyPath)
	if err != nil {
		return nil, err
	}

	ca, err := os.ReadFile(sc.config.Controlplane.CaPath)
	if err != nil {
		return nil, err
	}

	capool := x509.NewCertPool()
	if !capool.AppendCertsFromPEM(ca) {
		return nil, fmt.Errorf("Failed to load ca crt from %s", sc.config.Controlplane.CaPath)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		RootCAs:      capool,
	}

	return credentials.NewTLS(tlsConfig), nil
}

func (sc *SchedulerClient) newConnection() (*grpc.ClientConn, error) {
	var creds credentials.TransportCredentials
	if sc.config.Controlplane == nil || sc.config.Controlplane.KeyPath == "" {
		creds = insecure.NewCredentials()
	} else {
		tlsCreds, err := sc.loadKeyPair()
		if err != nil {
			return nil, err
		}
		creds = tlsCreds
	}

	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond)),
	}
	streamInterceptor := grpc_retry.StreamClientInterceptor(retryOpts...)
	unaryInterceptor := grpc_retry.UnaryClientInterceptor(retryOpts...)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithStreamInterceptor(streamInterceptor),
		grpc.WithUnaryInterceptor(unaryInterceptor),
		grpc.WithAuthority(sc.authority),
	}

	conn, err := grpc.Dial(sc.schedulerHost, opts...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func printProto(msg proto.Message) {
	resJson, err := protojson.Marshal(msg)
	if err != nil {
		fmt.Printf("Failed to print proto: %s", err.Error())
	} else {
		fmt.Printf("%s\n", string(resJson))
	}
}

func printProtoWithKey(key []byte, msg proto.Message) {
	resJson, err := protojson.Marshal(msg)
	if err != nil {
		fmt.Printf("Failed to print proto: %s", err.Error())
	} else {
		fmt.Printf("%s:%s\n", string(key), string(resJson))
	}
}

func unMarshallYamlStrict(data []byte, msg interface{}) error {
	jsonData, err := yaml.YAMLToJSON(data)
	if err != nil {
		return err
	}
	d := json.NewDecoder(bytes.NewReader(jsonData))
	d.DisallowUnknownFields() // So we fail if not exactly as required in schema
	err = d.Decode(msg)
	if err != nil {
		return err
	}
	return nil
}

func (sc *SchedulerClient) LoadModel(data []byte) (*scheduler.LoadModelResponse, error) {
	model := &mlopsv1alpha1.Model{}
	err := unMarshallYamlStrict(data, model)
	if err != nil {
		return nil, err
	}
	schModel, err := model.AsSchedulerModel()
	if err != nil {
		return nil, err
	}
	req := &scheduler.LoadModelRequest{Model: schModel}
	conn, err := sc.newConnection()
	if err != nil {
		return nil, err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)
	res, err := grpcClient.LoadModel(context.Background(), req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (sc *SchedulerClient) ListModels() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req := &scheduler.ModelStatusRequest{
		SubscriberName: subscriberName,
	}
	conn, err := sc.newConnection()
	if err != nil {
		return err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)

	stream, err := grpcClient.ModelStatus(ctx, req)
	if err != nil {
		return err
	}

	writer := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)
	_, err = fmt.Fprintln(writer, "model\tstate\treason")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(writer, "-----\t-----\t------")
	if err != nil {
		return err
	}
	for {
		res, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}

		}
		latestVersion := res.Versions[len(res.Versions)-1]
		if latestVersion.State.GetState() != scheduler.ModelStatus_ModelTerminated {
			_, err = fmt.Fprintf(writer, "%s\t%s\t%s\n", res.ModelName, latestVersion.State.GetState().String(), latestVersion.State.Reason)
			if err != nil {
				return err
			}
		}
	}
	err = writer.Flush()
	if err != nil {
		return err
	}
	return nil
}

func (sc *SchedulerClient) ModelStatus(modelName string, waitCondition string, timeoutSec int64) (*scheduler.ModelStatusResponse, error) {
	req := &scheduler.ModelStatusRequest{
		SubscriberName: subscriberName,
		Model: &scheduler.ModelReference{
			Name: modelName,
		},
	}

	conn, err := sc.newConnection()
	if err != nil {
		return nil, err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)

	var res *scheduler.ModelStatusResponse
	if waitCondition != "" {
		secsStart := time.Now().Unix()
		for {
			res, err := sc.getModelStatus(grpcClient, req)
			if err != nil {
				return nil, err
			}

			if len(res.Versions) > 0 {
				modelStatus := res.Versions[0].State.GetState().String()
				if modelStatus == waitCondition {
					break
				}
			}
			time.Sleep(1 * time.Second)
			if time.Now().Unix()-secsStart > timeoutSec {
				return nil, fmt.Errorf("Model wait status timeout")
			}
		}
	} else {
		res, err = sc.getModelStatus(grpcClient, req)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (sc *SchedulerClient) getModelStatus(
	grpcClient scheduler.SchedulerClient,
	req *scheduler.ModelStatusRequest,
) (*scheduler.ModelStatusResponse, error) {
	// There should only be one result, but cancel to ensure resources cleaned are up
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := grpcClient.ModelStatus(ctx, req)
	if err != nil {
		return nil, err
	}

	res, err := stream.Recv()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (sc *SchedulerClient) ServerStatus(serverName string, showRequest bool, showResponse bool) error {
	req := &scheduler.ServerStatusRequest{
		SubscriberName: subscriberName,
		Name:           &serverName,
	}
	if showRequest {
		printProto(req)
	}

	conn, err := sc.newConnection()
	if err != nil {
		return err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)

	res, err := sc.getServerStatus(grpcClient, req)
	if err != nil {
		return err
	}
	if !showResponse {
		fmt.Printf("%s loaded models %d available replicas %d\n", res.ServerName, res.NumLoadedModelReplicas, res.AvailableReplicas)
	} else {
		printProto(res)
	}
	return nil
}

func (sc *SchedulerClient) getServerStatus(
	grpcClient scheduler.SchedulerClient,
	req *scheduler.ServerStatusRequest,
) (*scheduler.ServerStatusResponse, error) {
	// There should only be one result, but cancel to ensure resources cleaned are up
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := grpcClient.ServerStatus(ctx, req)
	if err != nil {
		return nil, err
	}

	res, err := stream.Recv()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (sc *SchedulerClient) ListServers() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req := &scheduler.ServerStatusRequest{
		SubscriberName: subscriberName,
	}
	conn, err := sc.newConnection()
	if err != nil {
		return err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)

	stream, err := grpcClient.ServerStatus(ctx, req)
	if err != nil {
		return err
	}

	writer := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)
	_, err = fmt.Fprintln(writer, "server\treplicas\tmodels")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(writer, "------\t--------\t------")
	if err != nil {
		return err
	}
	for {
		res, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}

		}

		_, err = fmt.Fprintf(writer, "%s\t%d\t%d\n", res.ServerName, res.AvailableReplicas, res.NumLoadedModelReplicas)
		if err != nil {
			return err
		}
	}
	err = writer.Flush()
	if err != nil {
		return err
	}
	return nil
}

func (sc *SchedulerClient) UnloadModel(modelName string, modelBytes []byte) (*scheduler.UnloadModelResponse, error) {
	if len(modelBytes) > 0 && modelName == "" {
		model := &mlopsv1alpha1.Model{}
		err := unMarshallYamlStrict(modelBytes, model)
		if err != nil {
			return nil, err
		}
		modelName = model.Name
	}
	req := &scheduler.UnloadModelRequest{
		Model: &scheduler.ModelReference{
			Name: modelName,
		},
	}
	conn, err := sc.newConnection()
	if err != nil {
		return nil, err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)
	res, err := grpcClient.UnloadModel(context.Background(), req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (sc *SchedulerClient) StartExperiment(data []byte) (*scheduler.StartExperimentResponse, error) {
	experiment := &mlopsv1alpha1.Experiment{}
	err := unMarshallYamlStrict(data, experiment)
	if err != nil {
		return nil, err
	}
	schExperiment := experiment.AsSchedulerExperimentRequest()
	if err != nil {
		return nil, err
	}
	req := &scheduler.StartExperimentRequest{Experiment: schExperiment}
	conn, err := sc.newConnection()
	if err != nil {
		return nil, err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)
	res, err := grpcClient.StartExperiment(context.Background(), req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (sc *SchedulerClient) StopExperiment(experimentName string, experimentBytes []byte) (*scheduler.StopExperimentResponse, error) {
	if len(experimentBytes) > 0 && experimentName == "" {
		experiment := &mlopsv1alpha1.Experiment{}
		err := unMarshallYamlStrict(experimentBytes, experiment)
		if err != nil {
			return nil, err
		}
		experimentName = experiment.Name
	}
	req := &scheduler.StopExperimentRequest{
		Name: experimentName,
	}
	conn, err := sc.newConnection()
	if err != nil {
		return nil, err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)
	res, err := grpcClient.StopExperiment(context.Background(), req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (sc *SchedulerClient) ExperimentStatus(experimentName string, wait bool, timeoutSec int64) (*scheduler.ExperimentStatusResponse, error) {
	req := &scheduler.ExperimentStatusRequest{
		SubscriberName: subscriberName,
		Name:           &experimentName,
	}

	conn, err := sc.newConnection()
	if err != nil {
		return nil, err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)

	var res *scheduler.ExperimentStatusResponse
	if wait {
		secsStart := time.Now().Unix()
		for {
			res, err = sc.getExperimentStatus(grpcClient, req)
			if err != nil {
				return nil, err
			}
			if res.Active {
				break
			}
			time.Sleep(1 * time.Second)
			if time.Now().Unix()-secsStart > timeoutSec {
				return nil, fmt.Errorf("Experiment wait status timeout")
			}
		}
	} else {
		res, err = sc.getExperimentStatus(grpcClient, req)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (sc *SchedulerClient) getExperimentStatus(
	grpcClient scheduler.SchedulerClient,
	req *scheduler.ExperimentStatusRequest,
) (*scheduler.ExperimentStatusResponse, error) {
	// There should only be one result, but cancel to ensure resources cleaned are up
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := grpcClient.ExperimentStatus(ctx, req)
	if err != nil {
		return nil, err
	}

	res, err := stream.Recv()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (sc *SchedulerClient) ListExperiments() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req := &scheduler.ExperimentStatusRequest{
		SubscriberName: subscriberName,
	}
	conn, err := sc.newConnection()
	if err != nil {
		return err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)

	stream, err := grpcClient.ExperimentStatus(ctx, req)
	if err != nil {
		return err
	}

	writer := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)
	_, err = fmt.Fprintln(writer, "experiment\tactive\t")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(writer, "----------\t------\t")
	if err != nil {
		return err
	}
	for {
		res, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}

		}

		_, err = fmt.Fprintf(writer, "%s\t%v\n", res.ExperimentName, res.Active)
		if err != nil {
			return err
		}
	}
	err = writer.Flush()
	if err != nil {
		return err
	}
	return nil
}

func (sc *SchedulerClient) LoadPipeline(data []byte) (*scheduler.LoadPipelineResponse, error) {
	pipeline := &mlopsv1alpha1.Pipeline{}
	err := unMarshallYamlStrict(data, pipeline)
	if err != nil {
		return nil, err
	}
	schPipeline := pipeline.AsSchedulerPipeline()
	if err != nil {
		return nil, err
	}
	req := &scheduler.LoadPipelineRequest{Pipeline: schPipeline}
	conn, err := sc.newConnection()
	if err != nil {
		return nil, err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)
	res, err := grpcClient.LoadPipeline(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (sc *SchedulerClient) UnloadPipeline(pipelineName string, pipelineBytes []byte) (*scheduler.UnloadPipelineResponse, error) {

	if len(pipelineBytes) > 0 && pipelineName == "" {
		pipeline := &mlopsv1alpha1.Pipeline{}
		err := unMarshallYamlStrict(pipelineBytes, pipeline)
		if err != nil {
			return nil, err
		}
		pipelineName = pipeline.Name
	}

	req := &scheduler.UnloadPipelineRequest{
		Name: pipelineName,
	}
	conn, err := sc.newConnection()
	if err != nil {
		return nil, err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)
	res, err := grpcClient.UnloadPipeline(context.Background(), req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (sc *SchedulerClient) PipelineStatus(pipelineName string, waitCondition string, timeoutSec int64) (*scheduler.PipelineStatusResponse, error) {
	req := &scheduler.PipelineStatusRequest{
		SubscriberName: subscriberName,
		Name:           &pipelineName,
	}

	conn, err := sc.newConnection()
	if err != nil {
		return nil, err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)

	var res *scheduler.PipelineStatusResponse
	if waitCondition != "" {
		secsStart := time.Now().Unix()
		for {
			res, err = sc.getPipelineStatus(grpcClient, req)
			if err != nil {
				return nil, err
			}
			if len(res.Versions) > 0 {
				modelStatus := res.Versions[0].GetState().GetStatus().String()
				if modelStatus == waitCondition {
					break
				}
			}
			time.Sleep(1 * time.Second)
			if time.Now().Unix()-secsStart > timeoutSec {
				return nil, fmt.Errorf("Pipeline wait status timeout")
			}
		}
	} else {
		res, err = sc.getPipelineStatus(grpcClient, req)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (sc *SchedulerClient) getPipelineStatus(
	grpcClient scheduler.SchedulerClient,
	req *scheduler.PipelineStatusRequest,
) (*scheduler.PipelineStatusResponse, error) {
	// There should only be one result, but cancel to ensure resources cleaned are up
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := grpcClient.PipelineStatus(ctx, req)
	if err != nil {
		return nil, err
	}

	res, err := stream.Recv()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (sc *SchedulerClient) ListPipelines() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req := &scheduler.PipelineStatusRequest{
		SubscriberName: subscriberName,
	}
	conn, err := sc.newConnection()
	if err != nil {
		return err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)

	stream, err := grpcClient.PipelineStatus(ctx, req)
	if err != nil {
		return err
	}

	writer := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)
	_, err = fmt.Fprintln(writer, "pipeline\tstate\treason")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(writer, "--------\t-----\t------")
	if err != nil {
		return err
	}
	for {
		res, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}

		}
		pv := res.Versions[len(res.Versions)-1]
		_, err = fmt.Fprintf(writer, "%s\t%s\t%s\n", res.PipelineName, pv.State.Status.String(), pv.State.Reason)
		if err != nil {
			return err
		}
	}
	err = writer.Flush()
	if err != nil {
		return err
	}
	return nil
}
