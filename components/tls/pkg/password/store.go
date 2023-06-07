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

package password

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

type funcTLSServerOption struct {
	f func(options *PasswordStoreOptions)
}

func (fdo *funcTLSServerOption) apply(do *PasswordStoreOptions) {
	fdo.f(do)
}

func newFuncServerOption(f func(options *PasswordStoreOptions)) *funcTLSServerOption {
	return &funcTLSServerOption{
		f: f,
	}
}

type PasswordStore interface {
	GetPassword() string
	Stop()
}

type PasswordStoreOption interface {
	apply(options *PasswordStoreOptions)
}

type PasswordStoreOptions struct {
	prefix         string
	locationSuffix string
	clientset      kubernetes.Interface
}

func (c PasswordStoreOptions) String() string {
	return fmt.Sprintf("prefix=%s clientset=%v",
		c.prefix, c.clientset)
}

func getDefaultPasswordStoreOptions() PasswordStoreOptions {
	return PasswordStoreOptions{}
}

func Prefix(prefix string) PasswordStoreOption {
	return newFuncServerOption(func(o *PasswordStoreOptions) {
		o.prefix = prefix
	})
}

func LocationSuffix(suffix string) PasswordStoreOption {
	return newFuncServerOption(func(o *PasswordStoreOptions) {
		o.locationSuffix = suffix
	})
}
func ClientSet(clientSet kubernetes.Interface) PasswordStoreOption {
	return newFuncServerOption(func(o *PasswordStoreOptions) {
		o.clientset = clientSet
	})
}

func NewPasswordStore(opt ...PasswordStoreOption) (PasswordStore, error) {
	opts := getDefaultPasswordStoreOptions()
	for _, o := range opt {
		o.apply(&opts)
	}
	logger := logrus.New().WithField("source", "PasswordStore")
	logger.Infof("Options:%s", opts.String())
	if secretName, ok := util.GetEnv(opts.prefix, envSecretSuffix); ok {
		logger.Infof("Starting new password k8s secret store for %s from secret %s", opts.prefix, secretName)
		namespace, ok := os.LookupEnv(envNamespace)
		if !ok {
			return nil, fmt.Errorf("Namespace env var %s not found and needed for secret TLS", envNamespace)
		}
		ps, err := NewPasswordSecretHandler(secretName, opts.clientset, namespace, opts.prefix, opts.locationSuffix, logger)
		if err != nil {
			return nil, err
		}
		err = ps.GetPasswordAndWatch()
		if err != nil {
			return nil, err
		}
		return ps, nil
	} else {
		logger.Infof("Starting new password folder store for prefix %s", opts.prefix)
		fs, err := NewPasswordFolderHandler(opts.prefix, opts.locationSuffix, logger)
		if err != nil {
			return nil, err
		}
		err = fs.GetPasswordAndWatch()
		if err != nil {
			return nil, err
		}
		return fs, nil
	}
}
