package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operatorv2/apis/mlops/v1alpha1"

	"github.com/ghodss/yaml"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/seldonio/seldon-core/operatorv2/scheduler/apis/mlops/scheduler"
	"google.golang.org/grpc"
)

type SchedulerClient struct {
	schedulerHost string
	schedulerPort int
	callOptions   []grpc.CallOption
}

func NewSchedulerClient(schedulerHost string, schedulerPort int) *SchedulerClient {
	opts := []grpc.CallOption{
		grpc.MaxCallSendMsgSize(math.MaxInt32),
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
	}
	return &SchedulerClient{
		schedulerHost: schedulerHost,
		schedulerPort: schedulerPort,
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
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", sc.schedulerHost, sc.schedulerPort), opts...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (sc *SchedulerClient) LoadModel(data []byte, verbose bool) error {
	if verbose {
		fmt.Printf("%s\n", string(data))
	}
	model := &mlopsv1alpha1.Model{}
	err := yaml.Unmarshal(data, model)
	if err != nil {
		return err
	}
	schModel, err := model.AsSchedulerModel()
	if err != nil {
		return err
	}
	req := &scheduler.LoadModelRequest{Model: schModel}
	conn, err := sc.getConnection()
	if err != nil {
		return err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)
	_, err = grpcClient.LoadModel(context.Background(), req)
	if err != nil {
		return err
	}
	if verbose {
		fmt.Printf("Scheduled model %s for load\n", req.Model.Meta.Name)
	}
	return nil
}

func (sc *SchedulerClient) ModelStatus(modelName string, verbose bool, waitCondition string) error {
	req := &scheduler.ModelStatusRequest{
		Model: &scheduler.ModelReference{
			Name: modelName,
		},
	}
	conn, err := sc.getConnection()
	if err != nil {
		return err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)
	if waitCondition != "" {
		for {
			res, err := grpcClient.ModelStatus(context.Background(), req)
			if err != nil {
				return err
			}
			if len(res.Versions) > 0 {
				modelStatus := res.Versions[0].State.GetState().String()
				if modelStatus == waitCondition {
					fmt.Printf("{\"%s\":\"%s\"}\n", modelName, modelStatus)
					break
				}
			}
			time.Sleep(1 * time.Second)
		}
	} else {
		res, err := grpcClient.ModelStatus(context.Background(), req)
		if err != nil {
			return err
		}
		if !verbose {
			if len(res.Versions) > 0 {
				modelStatus := res.Versions[0].State.GetState().String()
				fmt.Printf("{\"%s\":\"%s\"}\n", modelName, modelStatus)
			} else {
				fmt.Println("Unknown")
			}
		} else {
			resBytesPretty, err := json.MarshalIndent(res, "", "    ")
			if err != nil {
				return err
			}
			fmt.Printf("%s\n", string(resBytesPretty))
		}
	}
	return nil
}

func (sc *SchedulerClient) ServerStatus(serverName string, verbose bool) error {
	req := &scheduler.ServerReference{
		Name: serverName,
	}
	conn, err := sc.getConnection()
	if err != nil {
		return err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)
	res, err := grpcClient.ServerStatus(context.Background(), req)
	if err != nil {
		return err
	}
	if !verbose {
		fmt.Printf("%s loaded models %d available replicas %d\n", res.ServerName, res.NumLoadedModelReplicas, res.AvailableReplicas)
	} else {
		resBytesPretty, err := json.MarshalIndent(res, "", "    ")
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", string(resBytesPretty))
	}
	return nil
}

func (sc *SchedulerClient) UnloadModel(modelName string, verbose bool) error {
	req := &scheduler.UnloadModelRequest{
		Model: &scheduler.ModelReference{
			Name: modelName,
		},
	}
	conn, err := sc.getConnection()
	if err != nil {
		return err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)
	_, err = grpcClient.UnloadModel(context.Background(), req)
	if err != nil {
		return err
	}
	if verbose {
		fmt.Printf("Scheduled model %s for unload\n", req.Model.Name)
	}
	return nil
}

func (sc *SchedulerClient) StartExperiment(data []byte, verbose bool) error {
	if verbose {
		fmt.Printf("%s\n", string(data))
	}
	experiment := &mlopsv1alpha1.Experiment{}
	err := yaml.Unmarshal(data, experiment)
	if err != nil {
		return err
	}
	schExperiment := experiment.AsSchedulerExperimentRequest()
	if err != nil {
		return err
	}
	req := &scheduler.StartExperimentRequest{Experiment: schExperiment}
	conn, err := sc.getConnection()
	if err != nil {
		return err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)
	_, err = grpcClient.StartExperiment(context.Background(), req)
	if err != nil {
		return err
	}
	if verbose {
		fmt.Printf("Scheduled experiment %s\n", schExperiment.Name)
	}
	return nil
}

func (sc *SchedulerClient) StopExperiment(experimentName string, verbose bool) error {
	req := &scheduler.StopExperimentRequest{
		Name: experimentName,
	}
	conn, err := sc.getConnection()
	if err != nil {
		return err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)
	_, err = grpcClient.StopExperiment(context.Background(), req)
	if err != nil {
		return err
	}
	if verbose {
		fmt.Printf("Scheduled experiment %s to stop\n", req.Name)
	}
	return nil
}

func (sc *SchedulerClient) ExperimentStatus(experimentName string, verbose bool, wait bool) error {
	req := &scheduler.ExperimentStatusRequest{
		Name: experimentName,
	}
	conn, err := sc.getConnection()
	if err != nil {
		return err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)
	var res *scheduler.ExperimentStatusResponse
	if wait {
		for {
			res, err = grpcClient.ExperimentStatus(context.Background(), req)
			if err != nil {
				return err
			}
			if res.Active {
				break
			}
			time.Sleep(1 * time.Second)
		}
	} else {
		res, err = grpcClient.ExperimentStatus(context.Background(), req)
		if err != nil {
			return err
		}
	}
	if verbose {
		resBytesPretty, err := json.MarshalIndent(res, "", "    ")
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", string(resBytesPretty))
	} else {
		fmt.Printf("%v", res.Active)
	}
	return nil
}

func (sc *SchedulerClient) LoadPipeline(data []byte, verbose bool) error {
	if verbose {
		fmt.Printf("%s\n", string(data))
	}
	pipeline := &mlopsv1alpha1.Pipeline{}
	err := yaml.Unmarshal(data, pipeline)
	if err != nil {
		return err
	}
	schPipeline := pipeline.AsSchedulerPipeline()
	if err != nil {
		return err
	}
	req := &scheduler.LoadPipelineRequest{Pipeline: schPipeline}
	conn, err := sc.getConnection()
	if err != nil {
		return err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)
	_, err = grpcClient.LoadPipeline(context.Background(), req)
	if err != nil {
		return err
	}
	if verbose {
		fmt.Printf("Scheduled pipeline %s for load\n", req.Pipeline.Name)
	}
	return nil
}

func (sc *SchedulerClient) UnloadPipeline(pipelineName string, verbose bool) error {
	req := &scheduler.UnloadPipelineRequest{
		Name: pipelineName,
	}
	conn, err := sc.getConnection()
	if err != nil {
		return err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)
	_, err = grpcClient.UnloadPipeline(context.Background(), req)
	if err != nil {
		return err
	}
	if verbose {
		fmt.Printf("Scheduled pipeline %s for unload\n", req.Name)
	}
	return nil
}

func (sc *SchedulerClient) PipelineStatus(pipelineName string, verbose bool) error {
	req := &scheduler.PipelineStatusRequest{
		Name: pipelineName,
	}
	conn, err := sc.getConnection()
	if err != nil {
		return err
	}
	grpcClient := scheduler.NewSchedulerClient(conn)
	res, err := grpcClient.PipelineStatus(context.Background(), req)
	if err != nil {
		return err
	}
	if verbose {
		resBytesPretty, err := json.MarshalIndent(res, "", "    ")
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", string(resBytesPretty))
	} else {
		if len(res.Versions) > 0 {
			fmt.Printf("%v", res.Versions[0].State.Status.String())
		} else {
			fmt.Println("Unknown status")
		}

	}
	return nil
}
