package k8sclient

import mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"

type Client interface {
	ApplyModel(model *mlopsv1alpha1.Model) error
	GetModel(model string) (*mlopsv1alpha1.Model, error)
	ApplyPipeline(pipeline *mlopsv1alpha1.Pipeline) error
	GetPipeline(pipeline string) (*mlopsv1alpha1.Pipeline, error)
}
