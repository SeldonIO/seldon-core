/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

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
	GetResources() []client.Object
	GetConditions() []*apis.Condition
}

func ToMetaObjects(objs []client.Object) []metav1.Object {
	var metaObjs []metav1.Object
	for _, res := range objs {
		metaObjs = append(metaObjs, res)
	}
	return metaObjs
}

func CopyMap[K, V comparable](m map[K]V) map[K]V {
	result := make(map[K]V)
	for k, v := range m {
		result[k] = v
	}
	return result
}

type ReplicaHandler interface {
	GetReplicas() (int32, error)
}

type LabelHandler interface {
	GetLabelSelector() string
}

type ReconcilerConfig struct {
	Ctx    context.Context
	Logger logr.Logger
	Client client.Client
}
