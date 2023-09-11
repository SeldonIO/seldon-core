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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func unMarshallYamlStrict(data []byte, msg interface{}) error {
	jsonData, err := yaml.YAMLToJSON(data)
	if err != nil {
		return err
	}
	d := json.NewDecoder(bytes.NewReader(jsonData))
	d.DisallowUnknownFields() // So we fail if not exactly as required in schema
	err = d.Decode(msg)
	if err != nil {
		return err
	}
	return nil
}

func moveStringDataToData(secret *v1.Secret) {
	secret.Data = make(map[string][]byte)
	for key, val := range secret.StringData {
		secret.Data[key] = []byte(val)
	}
}

func TestNewOAuthStoreWithSecret(t *testing.T) {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cc-oauth-test-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			fieldMethod:       []byte("OIDC"),
			fieldClientID:     []byte("test-client-id"),
			fieldClientSecret: []byte("test-client-secret"),
			fieldTokenURL:     []byte("https://example.com/openid-connect/token"),
			fieldExtensions:   []byte("logicalCluster=logic-1234,identityPoolId=pool-1234"),
			fieldScope:        []byte("test_scope"),
		},
	}

	prefix := "prefix"

	t.Setenv(fmt.Sprintf("%s%s", prefix, envSecretSuffix), secret.Name)
	t.Setenv(envNamespace, secret.Namespace)

	clientset := fake.NewSimpleClientset(secret)
	store, err := NewOAuthStore(Prefix(prefix), ClientSet(clientset))
	assert.NoError(t, err)

	oauthConfig := store.GetOAuthConfig()
	assert.Equal(t, "OIDC", oauthConfig.Method)
	assert.Equal(t, "test-client-id", oauthConfig.ClientID)
	assert.Equal(t, "test-client-secret", oauthConfig.ClientSecret)
	assert.Equal(t, "test_scope", oauthConfig.Scope)
	assert.Equal(t, "logicalCluster=logic-1234,identityPoolId=pool-1234", oauthConfig.Extensions)
	assert.Equal(t, "https://example.com/openid-connect/token", oauthConfig.TokenEndpointURL)

	newClientID := "new-client-id"
	secret.Data[fieldClientID] = []byte(newClientID)

	_, err = clientset.CoreV1().Secrets(secret.Namespace).Update(context.Background(), secret, metav1.UpdateOptions{})
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 500)

	oauthConfig = store.GetOAuthConfig()
	assert.Equal(t, newClientID, oauthConfig.ClientID)
	store.Stop()
}
