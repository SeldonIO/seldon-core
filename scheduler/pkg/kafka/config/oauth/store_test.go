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
	"os"
	"testing"
	"time"

	"github.com/ghodss/yaml"
	. "github.com/onsi/gomega"
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
	g := NewGomegaWithT(t)
	secretData, err := os.ReadFile("testdata/k8s_secret.yaml")
	g.Expect(err).To(BeNil())

	secret := &v1.Secret{}
	err = unMarshallYamlStrict(secretData, secret)
	g.Expect(err).To(BeNil())

	moveStringDataToData(secret)

	prefix := "prefix"

	t.Setenv(fmt.Sprintf("%s%s", prefix, envSecretSuffix), secret.Name)
	t.Setenv(envNamespace, secret.Namespace)

	clientset := fake.NewSimpleClientset(secret)
	ps, err := NewOAuthStore(Prefix(prefix), ClientSet(clientset))
	g.Expect(err).To(BeNil())

	oauthConfig := ps.GetOAuthConfig()
	g.Expect(oauthConfig.Method).To(Equal("OIDC"))
	g.Expect(oauthConfig.ClientID).To(Equal("test-client-id"))
	g.Expect(oauthConfig.ClientSecret).To(Equal("test-client-secret"))
	g.Expect(oauthConfig.Scope).To(Equal("test scope"))
	g.Expect(oauthConfig.TokenEndpointURL).To(Equal("https://keycloak.example.com/auth/realms/example-realm/protocol/openid-connect/token"))
	g.Expect(oauthConfig.Extensions).To(Equal("logicalCluster=logic-1234,identityPoolId=pool-1234"))

	newClientID := "new-client-id"
	secret.Data["client_id"] = []byte(newClientID)

	_, err = clientset.CoreV1().Secrets(secret.Namespace).Update(context.Background(), secret, metav1.UpdateOptions{})
	g.Expect(err).To(BeNil())
	time.Sleep(time.Millisecond * 500)

	oauthConfig = ps.GetOAuthConfig()
	g.Expect(oauthConfig.ClientID).To(Equal("new-client-id"))
	ps.Stop()
}
