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

type OAuthStore interface {
	GetOAuthConfig() OAuthConfig
	Stop()
}

type OAuthStoreOptions struct {
	Prefix         string
	LocationSuffix string
	Clientset      kubernetes.Interface
}

func (c OAuthStoreOptions) String() string {
	return fmt.Sprintf("prefix=%s locationSuffix=%s clientset=%v",
		c.Prefix, c.LocationSuffix, c.Clientset)
}

func NewOAuthStore(opts OAuthStoreOptions) (OAuthStore, error) {
	logger := logrus.New().WithField("source", "OAuthStore")
	logger.WithField("options", opts.String()).Info("creating new store")

	if secretName, ok := util.GetEnv(opts.Prefix, envSecretSuffix); ok {
		logger.
			WithField("secret", secretName).
			WithField("prefix", opts.Prefix).
			Info("creating store from secret")

		namespace, ok := os.LookupEnv(envNamespace)
		if !ok {
			return nil, fmt.Errorf("Namespace env var %s not found and needed for OAuth secret", envNamespace)
		}
		logger.WithField("namespace", namespace).Info("determined namespace from env var")

		store, err := NewOAuthSecretHandler(secretName, opts.Clientset, namespace, opts.Prefix, logger)
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
