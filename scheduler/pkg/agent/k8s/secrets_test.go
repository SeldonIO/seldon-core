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
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetSecretsConfig(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	type test struct {
		name       string
		secret     *v1.Secret
		secretName string
		err        bool
	}
	yamlTestData := `
type: s3                                                                                                  
name: s3                                                                                                  
parameters:                                                                                               
   provider: minio                                                                                         
   env_auth: false                                                                                         
   access_key_id: minioadmin                                                                               
   secret_access_key: minioadmin                                                                           
   endpoint: http://172.18.255.2:9000
`
	tests := []test{
		{
			name:       "simple",
			secret:     &v1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "mysec", Namespace: "default"}, Data: map[string][]byte{"mys3": []byte("xyz")}},
			secretName: "mysec",
			err:        false,
		},
		{
			name:       "NotFound",
			secret:     &v1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "mysec", Namespace: "default"}, Data: map[string][]byte{"mys3": []byte("xyz")}},
			secretName: "foo",
			err:        true,
		},
		{
			name:       "NoData",
			secret:     &v1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "mysec", Namespace: "default"}},
			secretName: "mysec",
			err:        true,
		},
		{
			name:       "StringData",
			secret:     &v1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "mysec", Namespace: "default"}, StringData: map[string]string{"mys3": yamlTestData}},
			secretName: "mysec",
			err:        false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakeClientset := fake.NewSimpleClientset(test.secret)
			s := NewSecretsHandler(fakeClientset, test.secret.Namespace)
			data, err := s.GetSecretConfig(test.secretName)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(data).ToNot(BeNil())
			}
		})
	}
}
