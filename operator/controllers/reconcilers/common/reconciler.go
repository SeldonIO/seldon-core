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

package common

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/go-logr/logr"
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
