/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package oauth

import (
	"context"
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
	clientset := fake.NewSimpleClientset(secret)

	prefix := "prefix"
	t.Setenv(prefix+envSecretSuffix, secret.Name)
	t.Setenv(envNamespace, secret.Namespace)

	store, err := NewOAuthStore(
		OAuthStoreOptions{
			Prefix:    prefix,
			Clientset: clientset,
		},
	)
	assert.NoError(t, err)
	defer store.Stop()

	t.Run(
		"get existing config",
		func(t *testing.T) {
			oauthConfig := store.GetOAuthConfig()
			assert.Equal(t, expectedExisting, oauthConfig)
		},
	)

	t.Run(
		"get updated config",
		func(t *testing.T) {
			secret.Data[fieldClientID] = []byte(expectedUpdate.ClientID)

			_, err = clientset.
				CoreV1().
				Secrets(secret.Namespace).
				Update(context.Background(), secret, metav1.UpdateOptions{})
			assert.NoError(t, err)

			checkForUpdate := func(c *assert.CollectT) {
				oauthConfig := store.GetOAuthConfig()
				assert.Equal(c, expectedUpdate, oauthConfig)
			}
			assert.EventuallyWithT(t, checkForUpdate, 100*time.Millisecond, 5*time.Millisecond)
		},
	)
}
