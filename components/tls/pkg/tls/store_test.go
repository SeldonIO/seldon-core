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

package tls

import (
	"context"
	"fmt"
	. "github.com/onsi/gomega"
	"github.com/otiai10/copy"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatchFolderCertificate(t *testing.T) {
	g := NewGomegaWithT(t)
	tmpFolder := t.TempDir()
	err := copy.Copy("testdata/cert1", tmpFolder)
	g.Expect(err).To(BeNil())
	prefix := "SCHEDULER"
	t.Setenv(fmt.Sprintf("%s%s", prefix, EnvCrtLocationSuffix), fmt.Sprintf("%s/tls.crt", tmpFolder))
	t.Setenv(fmt.Sprintf("%s%s", prefix, EnvKeyLocationSuffix), fmt.Sprintf("%s/tls.key", tmpFolder))
	t.Setenv(fmt.Sprintf("%s%s", prefix, EnvCaLocationSuffix), fmt.Sprintf("%s/ca.crt", tmpFolder))
	cs, err := NewCertificateStore(Prefix(prefix))
	g.Expect(err).To(BeNil())
	c1, err := cs.GetServerCertificate(nil)
	g.Expect(c1).ToNot(BeNil())
	g.Expect(err).To(BeNil())
	opt := copy.Options{
		OnDirExists: func(sr, dst string) copy.DirExistsAction { return copy.Replace },
		Sync:        true,
	}
	err = copy.Copy("testdata/cert2", tmpFolder, opt)
	g.Expect(err).To(BeNil())
	time.Sleep(time.Millisecond * 500)
	c2, err := cs.GetServerCertificate(nil)
	g.Expect(c2).ToNot(BeNil())
	g.Expect(err).To(BeNil())
	g.Expect(c1.Certificate).ToNot(Equal(c2.Certificate))
}

func TestWatchSecretCertificate(t *testing.T) {
	g := NewGomegaWithT(t)
	tlsKey, err := os.ReadFile("testdata/cert1/tls.key")
	g.Expect(err).To(BeNil())
	tlscrt, err := os.ReadFile("testdata/cert1/tls.crt")
	g.Expect(err).To(BeNil())
	ca, err := os.ReadFile("testdata/cert1/ca.crt")
	g.Expect(err).To(BeNil())
	sec := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "certs", Namespace: "default"},
		Data: map[string][]byte{
			"tls.key": tlsKey,
			"tls.crt": tlscrt,
			"ca.crt":  ca,
		},
	}
	prefix := "SCHEDULER"
	tmpFolder := t.TempDir()
	t.Setenv("SCHEDULER_TLS_SECRET_NAME", sec.Name)
	t.Setenv(fmt.Sprintf("%s%s", prefix, EnvCrtLocationSuffix), fmt.Sprintf("%s/tls.crt", tmpFolder))
	t.Setenv(fmt.Sprintf("%s%s", prefix, EnvKeyLocationSuffix), fmt.Sprintf("%s/tls.key", tmpFolder))
	t.Setenv(fmt.Sprintf("%s%s", prefix, EnvCaLocationSuffix), fmt.Sprintf("%s/ca.crt", tmpFolder))
	t.Setenv("POD_NAMESPACE", sec.Namespace)
	clientset := fake.NewSimpleClientset(&sec)
	cs, err := NewCertificateStore(Prefix(prefix), ClientSet(clientset))
	g.Expect(err).To(BeNil())
	c1, err := cs.GetServerCertificate(nil)
	g.Expect(c1).ToNot(BeNil())
	g.Expect(err).To(BeNil())
	tlsKey, err = os.ReadFile("testdata/cert2/tls.key")
	g.Expect(err).To(BeNil())
	tlscrt, err = os.ReadFile("testdata/cert2/tls.crt")
	g.Expect(err).To(BeNil())
	ca, err = os.ReadFile("testdata/cert2/ca.crt")
	g.Expect(err).To(BeNil())
	sec2 := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "certs", Namespace: "default"},
		Data: map[string][]byte{
			"tls.key": tlsKey,
			"tls.crt": tlscrt,
			"ca.crt":  ca,
		},
	}
	_, err = clientset.CoreV1().Secrets(sec.Namespace).Update(context.Background(), &sec2, metav1.UpdateOptions{})
	g.Expect(err).To(BeNil())
	time.Sleep(time.Millisecond * 500)
	c2, err := cs.GetServerCertificate(nil)
	g.Expect(c2).ToNot(BeNil())
	g.Expect(err).To(BeNil())
	g.Expect(c1.Certificate).ToNot(Equal(c2.Certificate))
	cs.Stop()
}

func TestWatchSecretCertificateWithValidation(t *testing.T) {
	g := NewGomegaWithT(t)
	tlsKey, err := os.ReadFile("testdata/cert1/tls.key")
	g.Expect(err).To(BeNil())
	tlscrt, err := os.ReadFile("testdata/cert1/tls.crt")
	g.Expect(err).To(BeNil())
	ca, err := os.ReadFile("testdata/cert1/ca.crt")
	g.Expect(err).To(BeNil())
	sec := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "certs", Namespace: "default"},
		Data: map[string][]byte{
			"tls.key": tlsKey,
			"tls.crt": tlscrt,
			"ca.crt":  ca,
		},
	}
	prefix := "SCHEDULER"
	tmpFolder := t.TempDir()
	path1 := filepath.Join(tmpFolder, "cert1")
	err = os.MkdirAll(path1, os.ModePerm)
	g.Expect(err).To(BeNil())
	t.Setenv("SCHEDULER_TLS_SECRET_NAME", sec.Name)
	t.Setenv(fmt.Sprintf("%s%s", prefix, EnvCrtLocationSuffix), fmt.Sprintf("%s/tls.crt", path1))
	t.Setenv(fmt.Sprintf("%s%s", prefix, EnvKeyLocationSuffix), fmt.Sprintf("%s/tls.key", path1))
	t.Setenv(fmt.Sprintf("%s%s", prefix, EnvCaLocationSuffix), fmt.Sprintf("%s/ca.crt", path1))

	secCa := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "certsca", Namespace: "default"},
		Data: map[string][]byte{
			"tls.key": tlsKey,
			"tls.crt": tlscrt,
			"ca.crt":  ca,
		},
	}
	path2 := filepath.Join(tmpFolder, "cert2")
	prefixCa := "SCHEDULER_CA"
	t.Setenv("SCHEDULER_CA_TLS_SECRET_NAME", secCa.Name)
	t.Setenv(fmt.Sprintf("%s%s", prefixCa, EnvCrtLocationSuffix), fmt.Sprintf("%s/tls.crt", path2))
	t.Setenv(fmt.Sprintf("%s%s", prefixCa, EnvKeyLocationSuffix), fmt.Sprintf("%s/tls.key", path2))
	t.Setenv(fmt.Sprintf("%s%s", prefixCa, EnvCaLocationSuffix), fmt.Sprintf("%s/ca.crt", path2))

	t.Setenv("POD_NAMESPACE", secCa.Namespace)
	clientset := fake.NewSimpleClientset(&sec, &secCa)
	cs, err := NewCertificateStore(Prefix(prefix), ClientSet(clientset), ValidationPrefix(prefixCa))
	g.Expect(err).To(BeNil())
	c1, err := cs.GetServerCertificate(nil)
	g.Expect(err).To(BeNil())
	g.Expect(c1).ToNot(BeNil())
	c := cs.GetCertificate()
	g.Expect(c).ToNot(BeNil())
	caCert := cs.GetValidationCertificate()
	g.Expect(caCert).ToNot(BeNil())
	cs.Stop()
}

func TestWatchValidationSecret(t *testing.T) {
	g := NewGomegaWithT(t)
	tlsKey, err := os.ReadFile("testdata/cert1/tls.key")
	g.Expect(err).To(BeNil())
	tlscrt, err := os.ReadFile("testdata/cert1/tls.crt")
	g.Expect(err).To(BeNil())
	ca, err := os.ReadFile("testdata/cert1/ca.crt")
	g.Expect(err).To(BeNil())

	tmpFolder := t.TempDir()
	path1 := filepath.Join(tmpFolder, "cert1")
	err = os.MkdirAll(path1, os.ModePerm)
	g.Expect(err).To(BeNil())

	secCa := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "certsca", Namespace: "default"},
		Data: map[string][]byte{
			"tls.key": tlsKey,
			"tls.crt": tlscrt,
			"ca.crt":  ca,
		},
	}
	path2 := filepath.Join(tmpFolder, "cert2")
	prefixCa := "SCHEDULER_CA"
	t.Setenv("SCHEDULER_CA_TLS_SECRET_NAME", secCa.Name)
	t.Setenv(fmt.Sprintf("%s%s", prefixCa, EnvCrtLocationSuffix), fmt.Sprintf("%s/tls.crt", path2))
	t.Setenv(fmt.Sprintf("%s%s", prefixCa, EnvKeyLocationSuffix), fmt.Sprintf("%s/tls.key", path2))
	t.Setenv(fmt.Sprintf("%s%s", prefixCa, EnvCaLocationSuffix), fmt.Sprintf("%s/ca.crt", path2))

	t.Setenv("POD_NAMESPACE", secCa.Namespace)
	clientset := fake.NewSimpleClientset(&secCa)
	cs, err := NewCertificateStore(ClientSet(clientset), ValidationPrefix(prefixCa), ValidationOnly(true))
	g.Expect(err).To(BeNil())
	caCert := cs.GetValidationCertificate()
	g.Expect(caCert).ToNot(BeNil())
	cs.Stop()
}
