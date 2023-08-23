/*
Copyright 2023 Seldon Technologies Ltd.

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

package oauth

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"

	"github.com/seldonio/seldon-core/components/tls/v2/pkg/util"
)

const (
	envSecretSuffix = "_OAUTH_SECRET_NAME"
	envNamespace    = "POD_NAMESPACE"
)

type funcOAUTHServerOption struct {
	f func(options *OAUTHStoreOptions)
}

func (fdo *funcOAUTHServerOption) apply(do *OAUTHStoreOptions) {
	fdo.f(do)
}

func newFuncServerOption(f func(options *OAUTHStoreOptions)) *funcOAUTHServerOption {
	return &funcOAUTHServerOption{
		f: f,
	}
}

type OAUTHStore interface {
	GetOAUTHConfig() OAUTHConfig
	Stop()
}

type OAUTHStoreOption interface {
	apply(options *OAUTHStoreOptions)
}

type OAUTHStoreOptions struct {
	prefix         string
	locationSuffix string
	clientset      kubernetes.Interface
}

func (c OAUTHStoreOptions) String() string {
	return fmt.Sprintf("prefix=%s locationSuffix=%s clientset=%v",
		c.prefix, c.locationSuffix, c.clientset)
}

func getDefaultOAUTHStoreOptions() OAUTHStoreOptions {
	return OAUTHStoreOptions{}
}

func Prefix(prefix string) OAUTHStoreOption {
	return newFuncServerOption(func(o *OAUTHStoreOptions) {
		o.prefix = prefix
	})
}

func LocationSuffix(suffix string) OAUTHStoreOption {
	return newFuncServerOption(func(o *OAUTHStoreOptions) {
		o.locationSuffix = suffix
	})
}
func ClientSet(clientSet kubernetes.Interface) OAUTHStoreOption {
	return newFuncServerOption(func(o *OAUTHStoreOptions) {
		o.clientset = clientSet
	})
}

func NewOAUTHStore(opt ...OAUTHStoreOption) (OAUTHStore, error) {
	opts := getDefaultOAUTHStoreOptions()
	for _, o := range opt {
		o.apply(&opts)
	}
	logger := logrus.New().WithField("source", "OAUTHStore")
	logger.Infof("Options:%s", opts.String())
	if secretName, ok := util.GetEnv(opts.prefix, envSecretSuffix); ok {
		logger.Infof("Starting new OAUTH k8s secret store for %s from secret %s", opts.prefix, secretName)
		namespace, ok := os.LookupEnv(envNamespace)
		logger.Infof("Namespace %s", namespace)
		if !ok {
			return nil, fmt.Errorf("Namespace env var %s not found and needed for OAUTH secret", envNamespace)
		}
		ps, err := NewOAUTHSecretHandler(secretName, opts.clientset, namespace, opts.prefix, logger)
		if err != nil {
			return nil, err
		}
		err = ps.GetOAUTHAndWatch()
		if err != nil {
			return nil, err
		}
		return ps, nil
	} else {
		// NOT IMPLEMENTED ERROR
		return nil, fmt.Errorf("OAUTH mechanism is currently only supported on K8s")
	}
}
