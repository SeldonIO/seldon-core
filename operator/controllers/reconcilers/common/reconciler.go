package common

import (
	"context"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler interface {
	Reconcile() error
	GetResources() []metav1.Object
	GetConditions() []*apis.Condition
}

type ReconcilerConfig struct {
	Ctx    context.Context
	Logger logr.Logger
	Client client.Client
}
