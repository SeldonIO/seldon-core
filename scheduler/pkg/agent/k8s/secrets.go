/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package k8s

import (
	"context"
	"fmt"
	"time"

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

func (s *SecretHandler) GetSecretConfig(secretName string, timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	secret, err := s.clientset.CoreV1().Secrets(s.namespace).Get(ctx, secretName, metav1.GetOptions{})
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
