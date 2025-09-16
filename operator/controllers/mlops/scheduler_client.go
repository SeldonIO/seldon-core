package mlops

import (
	"context"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
)

//go:generate go tool mockgen -source=./scheduler_client.go -destination=./mock/scheduler_client.go -package=mock SchedulerClient

type SchedulerClient interface {
	SubscribeControlPlaneEvents(ctx context.Context, grpcClient scheduler.SchedulerClient, namespace string) error
	StartExperiment(ctx context.Context, experiment *v1alpha1.Experiment, grpcClient scheduler.SchedulerClient) (bool, error)
	StopExperiment(ctx context.Context, experiment *v1alpha1.Experiment, grpcClient scheduler.SchedulerClient) (bool, error)
	SubscribeExperimentEvents(ctx context.Context, grpcClient scheduler.SchedulerClient, namespace string) error
	LoadModel(ctx context.Context, model *v1alpha1.Model, grpcClient scheduler.SchedulerClient) (bool, error)
	UnloadModel(ctx context.Context, model *v1alpha1.Model, grpcClient scheduler.SchedulerClient) (bool, error)
	SubscribeModelEvents(ctx context.Context, grpcClient scheduler.SchedulerClient, namespace string) error
	LoadPipeline(ctx context.Context, pipeline *v1alpha1.Pipeline, grpcClient scheduler.SchedulerClient) (bool, error)
	UnloadPipeline(ctx context.Context, pipeline *v1alpha1.Pipeline) (error, bool)
	SubscribePipelineEvents(ctx context.Context, grpcClient scheduler.SchedulerClient, namespace string) error
	ServerNotify(ctx context.Context, grpcClient scheduler.SchedulerClient, servers []v1alpha1.Server, isFirstSync bool) error
	SubscribeServerEvents(ctx context.Context, grpcClient scheduler.SchedulerClient, namespace string) error
	RemoveConnection(namespace string)
}
