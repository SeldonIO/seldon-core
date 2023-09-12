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
	envSecretSuffix = "_SASL_SECRET_NAME"
	envNamespace    = "POD_NAMESPACE"
)

type OAuthConfig struct {
	Method           string
	ClientID         string
	ClientSecret     string
	Scope            string
	TokenEndpointURL string
	Extensions       string
}

type funcOAuthServerOption struct {
	f func(options *OAuthStoreOptions)
}

func (fdo *funcOAuthServerOption) apply(do *OAuthStoreOptions) {
	fdo.f(do)
}

func newFuncServerOption(f func(options *OAuthStoreOptions)) *funcOAuthServerOption {
	return &funcOAuthServerOption{
		f: f,
	}
}

type OAuthStore interface {
	GetOAuthConfig() OAuthConfig
	Stop()
}

type OAuthStoreOption interface {
	apply(options *OAuthStoreOptions)
}

type OAuthStoreOptions struct {
	prefix         string
	locationSuffix string
	clientset      kubernetes.Interface
}

func (c OAuthStoreOptions) String() string {
	return fmt.Sprintf("prefix=%s locationSuffix=%s clientset=%v",
		c.prefix, c.locationSuffix, c.clientset)
}

func getDefaultOAuthStoreOptions() OAuthStoreOptions {
	return OAuthStoreOptions{}
}

func Prefix(prefix string) OAuthStoreOption {
	return newFuncServerOption(func(o *OAuthStoreOptions) {
		o.prefix = prefix
	})
}

func LocationSuffix(suffix string) OAuthStoreOption {
	return newFuncServerOption(func(o *OAuthStoreOptions) {
		o.locationSuffix = suffix
	})
}
func ClientSet(clientSet kubernetes.Interface) OAuthStoreOption {
	return newFuncServerOption(func(o *OAuthStoreOptions) {
		o.clientset = clientSet
	})
}

func NewOAuthStore(opt ...OAuthStoreOption) (OAuthStore, error) {
	opts := getDefaultOAuthStoreOptions()
	for _, o := range opt {
		o.apply(&opts)
	}

	logger := logrus.New().WithField("source", "OAuthStore")
	logger.Infof("Options:%s", opts.String())

	if secretName, ok := util.GetEnv(opts.prefix, envSecretSuffix); ok {
		logger.Infof("Starting new OAuth k8s secret store for %s from secret %s", opts.prefix, secretName)

		namespace, ok := os.LookupEnv(envNamespace)
		if !ok {
			return nil, fmt.Errorf("Namespace env var %s not found and needed for OAuth secret", envNamespace)
		}
		logger.Infof("Namespace %s", namespace)

		store, err := NewOAuthSecretHandler(secretName, opts.clientset, namespace, opts.prefix, logger)
		if err != nil {
			return nil, err
		}

		err = store.GetOAuthAndWatch()
		if err != nil {
			return nil, err
		}

		return store, nil
	} else {
		return nil, fmt.Errorf("OAuth mechanism is currently only supported on K8s")
	}
}
