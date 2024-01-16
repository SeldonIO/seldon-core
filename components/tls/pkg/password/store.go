/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package password

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

type PasswordStore interface {
	GetPassword() string
	Stop()
}

type PasswordStoreOptions struct {
	Prefix         string
	LocationSuffix string
	Clientset      kubernetes.Interface
}

func (c PasswordStoreOptions) String() string {
	return fmt.Sprintf("prefix=%s locationSuffix=%s clientset=%v",
		c.Prefix, c.LocationSuffix, c.Clientset)
}

func NewPasswordStore(opts PasswordStoreOptions) (PasswordStore, error) {
	secretName, ok := util.GetNonEmptyEnv(opts.Prefix, envSecretSuffix)
	if ok {
		return newPasswordStoreFromSecret(secretName, opts)
	} else {
		return newPasswordStoreFromFile(opts)
	}
}

func newPasswordStoreFromSecret(secretName string, opts PasswordStoreOptions) (PasswordStore, error) {
	baseLogger := logrus.New()
	logger := baseLogger.WithField("source", "PasswordStore")

	namespace, ok := util.GetNonEmptyEnv("", envNamespace)
	if !ok {
		return nil, fmt.Errorf("%s not found but required for password secret", envNamespace)
	}

	logger.
		WithField("namespace", namespace).
		WithField("secret", secretName).
		WithField("options", opts.String()).
		Info("creating store from secret")

	store, err := newK8sSecretStore(
		secretName,
		opts.Clientset,
		namespace,
		opts.Prefix,
		opts.LocationSuffix,
		baseLogger,
	)
	if err != nil {
		return nil, err
	}

	err = store.loadAndWatchPassword()
	if err != nil {
		return nil, err
	}

	return store, nil
}

func newPasswordStoreFromFile(opts PasswordStoreOptions) (PasswordStore, error) {
	baseLogger := logrus.New()
	logger := baseLogger.WithField("source", "PasswordStore")
	logger.
		WithField("options", opts.String()).
		Info("creating store from file")

	fs, err := newFileStore(opts.Prefix, opts.LocationSuffix, baseLogger)
	if err != nil {
		return nil, err
	}

	err = fs.loadAndWatchPassword()
	if err != nil {
		return nil, err
	}

	return fs, nil
}
