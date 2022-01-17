package scheduler

import (
	"context"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/seldonio/seldon-core/operatorv2/pkg/constants"
	"github.com/seldonio/seldon-core/operatorv2/pkg/utils"
	"k8s.io/client-go/util/retry"

	apimachinary_errors "k8s.io/apimachinery/pkg/api/errors"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/seldonio/seldon-core/operatorv2/apis/mlops/v1alpha1"
	mlopsv1alpha1 "github.com/seldonio/seldon-core/operatorv2/apis/mlops/v1alpha1"
	scheduler "github.com/seldonio/seldon-core/operatorv2/scheduler/apis/mlops/scheduler"
	"google.golang.org/grpc"
)

type SchedulerClient struct {
	client.Client
	logger      logr.Logger
	conn        *grpc.ClientConn
	callOptions []grpc.CallOption
	recorder    record.EventRecorder
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

func (s *SchedulerClient) ConnectToScheduler(host string, port int) error {
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond)),
	}
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(retryOpts...)),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOpts...)),
	}
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), opts...)
	if err != nil {
		return err
	}
	s.conn = conn
	s.logger.Info("Connected to scheduler")
	return nil
}

func (s *SchedulerClient) LoadModel(ctx context.Context, model *v1alpha1.Model) error {
	logger := s.logger.WithName("LoadModel")
	grcpClient := scheduler.NewSchedulerClient(s.conn)

	md, err := model.AsSchedulerModel()
	if err != nil {
		return err
	}
	loadModelRequest := scheduler.LoadModelRequest{
		Model: md,
	}

	logger.Info("Load", "model name", model.Name)
	_, err = grcpClient.LoadModel(ctx, &loadModelRequest, grpc_retry.WithMax(2))
	if err != nil {
		return err
	}
	return nil
}

func (s *SchedulerClient) UnloadModel(ctx context.Context, model *v1alpha1.Model) error {
	logger := s.logger.WithName("UnloadModel")
	grcpClient := scheduler.NewSchedulerClient(s.conn)

	modelRef := &scheduler.UnloadModelRequest{
		Model: &scheduler.ModelReference{
			Name: model.Name,
		},
		KubernetesMeta: &scheduler.KubernetesMeta{
			Namespace:  model.Namespace,
			Generation: model.Generation,
		},
	}
	logger.Info("Unload", "model name", model.Name)
	_, err := grcpClient.UnloadModel(ctx, modelRef, grpc_retry.WithMax(2))
	if err != nil {
		return err
	}
	return nil
}

func (s *SchedulerClient) SubscribeEvents(ctx context.Context) error {
	logger := s.logger.WithName("SubscribeEvent")
	grcpClient := scheduler.NewSchedulerClient(s.conn)

	stream, err := grcpClient.SubscribeModelStatus(ctx, &scheduler.ModelSubscriptionRequest{Name: "seldon manager"}, grpc_retry.WithMax(1))
	if err != nil {
		return err
	}
	for {
		event, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			logger.Error(err, "event recv failed")
			return err
		}
		// The expected contract is just the latest version will be sent to us
		if len(event.Versions) < 1 {
			logger.Info("Expected a single model version", "numVersions", len(event.Versions), "name", event.GetModelName())
			continue
		}
		latestVersionStatus := event.Versions[0]
		if latestVersionStatus.GetKubernetesMeta() == nil {
			logger.Info("Ignoring event with no Kubernetes metadata.", "model", event.ModelName)
			continue
		}
		logger.Info("Received event", "name", event.ModelName, "version", latestVersionStatus.Version, "generation", latestVersionStatus.GetKubernetesMeta().Generation, "state", latestVersionStatus.State.State.String(), "reason", latestVersionStatus.State.Reason)

		// Handle terminated event to remove finalizer
		if latestVersionStatus.State.State == scheduler.ModelStatus_ModelTerminated {
			retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				latestModel := &mlopsv1alpha1.Model{}
				err = s.Get(ctx, client.ObjectKey{Name: event.ModelName, Namespace: latestVersionStatus.GetKubernetesMeta().Namespace}, latestModel)
				if err != nil {
					return err
				}
				// remove finalizer now we have completed successfully
				latestModel.ObjectMeta.Finalizers = utils.RemoveStr(latestModel.ObjectMeta.Finalizers, constants.ModelFinalizerName)
				if err := s.Update(ctx, latestModel); err != nil {
					logger.Error(err, "Failed to remove finalizer", "model", latestModel.GetName())
					return err
				}
				return nil
			})
			if retryErr != nil {
				logger.Error(err, "Failed to remove finalizer after retries")
			}
			continue
		}

		// Try to update status
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			latestModel := &mlopsv1alpha1.Model{}
			err = s.Get(ctx, client.ObjectKey{Name: event.ModelName, Namespace: latestVersionStatus.GetKubernetesMeta().Namespace}, latestModel)
			if err != nil {
				return err
			}
			if latestVersionStatus.GetKubernetesMeta().Generation != latestModel.Generation {
				logger.Info("Ignoring event for old generation", "currentGeneration", latestModel.Generation, "eventGeneration", latestVersionStatus.GetKubernetesMeta().Generation, "model", event.ModelName)
				return nil
			}
			// Handle status update
			switch latestVersionStatus.State.State {
			case scheduler.ModelStatus_ModelAvailable:
				latestModel.Status.CreateAndSetCondition(mlopsv1alpha1.SeldonMeshReady, true, latestVersionStatus.State.Reason)
			default:
				latestModel.Status.CreateAndSetCondition(mlopsv1alpha1.SeldonMeshReady, false, latestVersionStatus.State.Reason)
			}
			return s.updateStatus(latestModel)
		})
		if retryErr != nil {
			logger.Error(err, "Failed to update status", "model", event.ModelName)
		}

	}
	return nil
}

func modelReady(status mlopsv1alpha1.ModelStatus) bool {
	return status.Conditions != nil &&
		status.GetCondition(apis.ConditionReady) != nil &&
		status.GetCondition(apis.ConditionReady).Status == v1.ConditionTrue
}

func (s *SchedulerClient) updateStatus(model *mlopsv1alpha1.Model) error {
	//logger := s.logger.WithName("updateStatus")
	model.Status.CreateAndSetCondition(mlopsv1alpha1.DeploymentsReady, true, "No deployments used")
	existingModel := &mlopsv1alpha1.Model{}
	namespacedName := types.NamespacedName{Name: model.Name, Namespace: model.Namespace}
	if err := s.Get(context.TODO(), namespacedName, existingModel); err != nil {
		if apimachinary_errors.IsNotFound(err) { //Ignore NotFound errors
			return nil
		}
		return err
	}
	prevWasReady := modelReady(existingModel.Status)
	if equality.Semantic.DeepEqual(existingModel.Status, model.Status) {
		// Not updating as no difference
	} else {
		if err := s.Status().Update(context.TODO(), model); err != nil {
			s.recorder.Eventf(model, v1.EventTypeWarning, "UpdateFailed",
				"Failed to update status for Model %q: %v", model.Name, err)
			return err
		} else {
			currentIsReady := modelReady(model.Status)
			if prevWasReady && !currentIsReady {
				s.recorder.Eventf(model, v1.EventTypeWarning, "ModelNotReady",
					fmt.Sprintf("Model [%v] is no longer Ready", model.GetName()))
			} else if !prevWasReady && currentIsReady {
				s.recorder.Eventf(model, v1.EventTypeNormal, "ModelReady",
					fmt.Sprintf("Model [%v] is Ready", model.GetName()))
			}
		}
	}
	return nil
}
