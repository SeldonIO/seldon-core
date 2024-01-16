/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package password

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"
)

func TestWatchSecretPassword(t *testing.T) {
	g := NewGomegaWithT(t)
	password, err := os.ReadFile("testdata/p1/password")
	g.Expect(err).To(BeNil())
	secName := "password"
	sec := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: secName, Namespace: "default"},
		Data: map[string][]byte{
			"password": password,
		},
	}
	prefix := tls.EnvSecurityPrefixKafkaClient
	suffix := "suffix"
	tmpFolder := t.TempDir()
	t.Setenv(fmt.Sprintf("%s%s", tls.EnvSecurityPrefixKafkaClient, envSecretSuffix), sec.Name)
	t.Setenv(fmt.Sprintf("%s%s", prefix, suffix), fmt.Sprintf("%s/password", tmpFolder))
	t.Setenv("POD_NAMESPACE", sec.Namespace)
	clientset := fake.NewSimpleClientset(&sec)
	ps, err := newK8sSecretStore(secName, clientset, "default", prefix, suffix, logrus.New())
	g.Expect(err).To(BeNil())
	err = ps.loadAndWatchPassword()
	g.Expect(err).To(BeNil())
	passwordLoaded := ps.GetPassword()
	g.Expect(passwordLoaded).To(Equal(string(password)))
	password = []byte("newpassword")
	sec2 := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: secName, Namespace: "default"},
		Data: map[string][]byte{
			"password": password,
		},
	}
	_, err = clientset.CoreV1().Secrets(sec.Namespace).Update(context.Background(), &sec2, metav1.UpdateOptions{})
	g.Expect(err).To(BeNil())
	time.Sleep(time.Millisecond * 500)
	passwordLoaded = ps.GetPassword()
	g.Expect(passwordLoaded).To(Equal(string(password)))
	ps.Stop()
}
