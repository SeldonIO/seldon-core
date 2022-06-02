package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"text/tabwriter"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operatorv2/apis/mlops/v1alpha1"

	"github.com/ghodss/yaml"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/seldonio/seldon-core/operatorv2/scheduler/apis/mlops/scheduler"
	"google.golang.org/grpc"
)

const subscriberName = "seldon CLI"

type SchedulerClient struct {
	schedulerHost string
	callOptions   []grpc.CallOption
}

func NewSchedulerClient(schedulerHost string) *SchedulerClient {

	opts := []grpc.CallOption{
		grpc.MaxCallSendMsgSize(math.MaxInt32),
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
	}
	return &SchedulerClient{
		schedulerHost: schedulerHost,
		callOptions:   opts,
	}
}

func (sc *SchedulerClient) getConnection() (*grpc.ClientConn, error) {
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond)),
	}
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(retryOpts...)),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOpts...)),
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

func (sc *SchedulerClient) LoadModel(data []byte, showRequest bool, showResponse bool) error {
	model := &mlopsv1alpha1.Model{}
	err := unMarshallYamlStrict(data, model)
	if err != nil {
		return err
	}
	schModel, err := model.AsSchedulerModel()
	if err != nil {
		return err
	}
	req := &scheduler.LoadModelRequest{Model: schModel}
	if showRequest {
		printProto(req)
	}
	conn, err := sc.getConnection()
	if err != nil {
		return err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)
	res, err := grpcClient.LoadModel(context.Background(), req)
	if err != nil {
		return err
	}
	if showResponse {
		printProto(res)
	}
	return nil
}

func (sc *SchedulerClient) ListModels() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req := &scheduler.ModelStatusRequest{
		SubscriberName: subscriberName,
	}
	conn, err := sc.getConnection()
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

func (sc *SchedulerClient) ModelStatus(modelName string, showRequest bool, showResponse bool, waitCondition string) error {
	req := &scheduler.ModelStatusRequest{
		SubscriberName: subscriberName,
		Model: &scheduler.ModelReference{
			Name: modelName,
		},
	}
	if showRequest {
		printProto(req)
	}

	conn, err := sc.getConnection()
	if err != nil {
		return err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)

	var res *scheduler.ModelStatusResponse
	if waitCondition != "" {
		for {
			res, err := sc.getModelStatus(grpcClient, req)
			if err != nil {
				return err
			}

			if len(res.Versions) > 0 {
				modelStatus := res.Versions[0].State.GetState().String()
				if modelStatus == waitCondition {
					break
				}
			}
			time.Sleep(1 * time.Second)
		}
	} else {
		res, err = sc.getModelStatus(grpcClient, req)
		if err != nil {
			return err
		}
	}
	if !showResponse {
		if len(res.Versions) > 0 {
			modelStatus := res.Versions[0].State.GetState().String()
			fmt.Printf("{\"%s\":\"%s\"}\n", modelName, modelStatus)
		} else {
			fmt.Println("Unknown")
		}
	} else {
		printProto(res)
	}
	return nil
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

	conn, err := sc.getConnection()
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
	conn, err := sc.getConnection()
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

func (sc *SchedulerClient) UnloadModel(modelName string, showRequest bool, showResponse bool) error {
	req := &scheduler.UnloadModelRequest{
		Model: &scheduler.ModelReference{
			Name: modelName,
		},
	}
	if showRequest {
		printProto(req)
	}
	conn, err := sc.getConnection()
	if err != nil {
		return err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)
	res, err := grpcClient.UnloadModel(context.Background(), req)
	if err != nil {
		return err
	}
	if showResponse {
		printProto(res)
	}
	return nil
}

func (sc *SchedulerClient) StartExperiment(data []byte, showRequest bool, showResponse bool) error {
	experiment := &mlopsv1alpha1.Experiment{}
	err := unMarshallYamlStrict(data, experiment)
	if err != nil {
		return err
	}
	schExperiment := experiment.AsSchedulerExperimentRequest()
	if err != nil {
		return err
	}
	req := &scheduler.StartExperimentRequest{Experiment: schExperiment}
	if showRequest {
		printProto(req)
	}
	conn, err := sc.getConnection()
	if err != nil {
		return err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)
	res, err := grpcClient.StartExperiment(context.Background(), req)
	if err != nil {
		return err
	}
	if showResponse {
		printProto(res)
	}
	return nil
}

func (sc *SchedulerClient) StopExperiment(experimentName string, showRequest bool, showResponse bool) error {
	req := &scheduler.StopExperimentRequest{
		Name: experimentName,
	}
	if showRequest {
		printProto(req)
	}
	conn, err := sc.getConnection()
	if err != nil {
		return err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)
	res, err := grpcClient.StopExperiment(context.Background(), req)
	if err != nil {
		return err
	}
	if showResponse {
		printProto(res)
	}
	return nil
}

func (sc *SchedulerClient) ExperimentStatus(experimentName string, showRequest bool, showResponse bool, wait bool) error {
	req := &scheduler.ExperimentStatusRequest{
		SubscriberName: subscriberName,
		Name:           &experimentName,
	}
	if showRequest {
		printProto(req)
	}

	conn, err := sc.getConnection()
	if err != nil {
		return err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)

	var res *scheduler.ExperimentStatusResponse
	if wait {
		for {
			res, err = sc.getExperimentStatus(grpcClient, req)
			if err != nil {
				return err
			}
			if res.Active {
				break
			}
			time.Sleep(1 * time.Second)
		}
	} else {
		res, err = sc.getExperimentStatus(grpcClient, req)
		if err != nil {
			return err
		}
	}
	if showResponse {
		printProto(res)
	} else {
		fmt.Printf("%v", res.Active)
	}
	return nil
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
	conn, err := sc.getConnection()
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

func (sc *SchedulerClient) LoadPipeline(data []byte, showRequest bool, showResponse bool) error {
	pipeline := &mlopsv1alpha1.Pipeline{}
	err := unMarshallYamlStrict(data, pipeline)
	if err != nil {
		return err
	}
	schPipeline := pipeline.AsSchedulerPipeline()
	if err != nil {
		return err
	}
	req := &scheduler.LoadPipelineRequest{Pipeline: schPipeline}
	if showRequest {
		printProto(req)
	}
	conn, err := sc.getConnection()
	if err != nil {
		return err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)
	res, err := grpcClient.LoadPipeline(context.Background(), req)
	if err != nil {
		return err
	}
	if showResponse {
		printProto(res)
	}
	return nil
}

func (sc *SchedulerClient) UnloadPipeline(pipelineName string, showRequest bool, showResponse bool) error {
	req := &scheduler.UnloadPipelineRequest{
		Name: pipelineName,
	}
	if showRequest {
		printProto(req)
	}
	conn, err := sc.getConnection()
	if err != nil {
		return err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)
	res, err := grpcClient.UnloadPipeline(context.Background(), req)
	if err != nil {
		return err
	}
	if showResponse {
		printProto(res)
	}
	return nil
}

func (sc *SchedulerClient) PipelineStatus(pipelineName string, showRequest bool, showResponse bool, waitCondition string) error {
	req := &scheduler.PipelineStatusRequest{
		SubscriberName: subscriberName,
		Name:           &pipelineName,
	}
	if showRequest {
		printProto(req)
	}

	conn, err := sc.getConnection()
	if err != nil {
		return err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)

	var res *scheduler.PipelineStatusResponse
	if waitCondition != "" {
		for {
			res, err = sc.getPipelineStatus(grpcClient, req)
			if err != nil {
				return err
			}
			if len(res.Versions) > 0 {
				modelStatus := res.Versions[0].GetState().GetStatus().String()
				if modelStatus == waitCondition {
					break
				}
			}
			time.Sleep(1 * time.Second)
		}
	} else {
		res, err = sc.getPipelineStatus(grpcClient, req)
		if err != nil {
			return err
		}
	}
	if showResponse {
		printProto(res)
	} else {
		if len(res.Versions) > 0 {
			fmt.Printf("%v", res.Versions[0].State.Status.String())
		} else {
			fmt.Println("Unknown status")
		}

	}
	return nil
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
	conn, err := sc.getConnection()
	if err != nil {
		return err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)

	stream, err := grpcClient.PipelineStatus(ctx, req)
	if err != nil {
		return err
	}

	writer := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)
	_, err = fmt.Fprintln(writer, "pipeline\tstate")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(writer, "--------\t-----")
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

		_, err = fmt.Fprintf(writer, "%s\t%s\n", res.PipelineName, res.Versions[len(res.Versions)-1].State.Status.String())
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
