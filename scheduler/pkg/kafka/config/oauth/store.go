/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package oauth

import (
	"fmt"

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
	secretName, ok := util.GetNonEmptyEnv(opts.Prefix, envSecretSuffix)
	if !ok {
		return nil, fmt.Errorf("OAuth mechanism is currently only supported on K8s")
	}

	namespace, ok := util.GetNonEmptyEnv("", envNamespace)
	if !ok {
		return nil, fmt.Errorf("%s not found but required for OAuth secret", envNamespace)
	}

	baseLogger := logrus.New()
	logger := baseLogger.WithField("source", "OAuthStore")
	logger.
		WithField("namespace", namespace).
		WithField("secret", secretName).
		WithField("options", opts.String()).
		Info("creating store from secret")

	store, err := newK8sSecretStore(secretName, opts.Clientset, namespace, opts.Prefix, baseLogger)
	if err != nil {
		return nil, err
	}

	err = store.loadAndWatchConfig()
	if err != nil {
		return nil, err
	}

	return store, nil
}
