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

package k8s

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type SecretHandler struct {
	clientset kubernetes.Interface
	namespace string
}

func NewSecretsHandler(clientset kubernetes.Interface, namespace string) *SecretHandler {
	return &SecretHandler{
		clientset: clientset,
		namespace: namespace,
	}
}

func (s *SecretHandler) GetSecretConfig(secretName string) ([]byte, error) {
	secret, err := s.clientset.CoreV1().Secrets(s.namespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if len(secret.Data) == 1 {
		for _, val := range secret.Data {
			return val, nil
		}
	}

	if len(secret.StringData) == 1 {
		for _, val := range secret.StringData {
			return []byte(val), nil
		}
	}

	//TODO allow more than 1 key in secret
	return nil, fmt.Errorf("Secret does not have 1 key %s", secretName)
}
