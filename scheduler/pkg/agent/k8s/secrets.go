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
	return nil, fmt.Errorf("Secret does not have 1 key %s", secretName)
}
