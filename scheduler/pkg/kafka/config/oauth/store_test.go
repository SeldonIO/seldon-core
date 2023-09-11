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
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewOAuthStoreWithSecret(t *testing.T) {
	expectedExisting := OAuthConfig{
		Method:           "OIDC",
		ClientID:         "test-client-id",
		ClientSecret:     "test-client-secret",
		TokenEndpointURL: "https://example.com/openid-connect/token",
		Extensions:       "logicalCluster=logic-1234,identityPoolId=pool-1234",
		Scope:            "test_scope",
	}
	expectedUpdate := expectedExisting
	expectedUpdate.ClientID = "new-client-id"

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cc-oauth-test-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			fieldMethod:       []byte(expectedExisting.Method),
			fieldClientID:     []byte(expectedExisting.ClientID),
			fieldClientSecret: []byte(expectedExisting.ClientSecret),
			fieldTokenURL:     []byte(expectedExisting.TokenEndpointURL),
			fieldExtensions:   []byte(expectedExisting.Extensions),
			fieldScope:        []byte(expectedExisting.Scope),
		},
	}

	prefix := "prefix"

	t.Setenv(fmt.Sprintf("%s%s", prefix, envSecretSuffix), secret.Name)
	t.Setenv(envNamespace, secret.Namespace)

	clientset := fake.NewSimpleClientset(secret)
	store, err := NewOAuthStore(Prefix(prefix), ClientSet(clientset))
	assert.NoError(t, err)

	oauthConfig := store.GetOAuthConfig()
	assert.Equal(t, expectedExisting, oauthConfig)

	secret.Data[fieldClientID] = []byte(expectedUpdate.ClientID)

	_, err = clientset.
		CoreV1().
		Secrets(secret.Namespace).
		Update(context.Background(), secret, metav1.UpdateOptions{})
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 500)

	oauthConfig = store.GetOAuthConfig()
	assert.Equal(t, expectedUpdate, oauthConfig)
	store.Stop()
}
