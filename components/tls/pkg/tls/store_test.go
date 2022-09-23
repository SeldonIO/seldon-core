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
	"testing"
	"time"
)

func TestWatchFolderCertificate(t *testing.T) {
	g := NewGomegaWithT(t)
	tmpFolder := t.TempDir()
	err := copy.Copy("testdata/cert1", tmpFolder)
	g.Expect(err).To(BeNil())
	prefix := "SCHEDULER"
	t.Setenv(fmt.Sprintf("%s%s", prefix, envCrtLocationSuffix), fmt.Sprintf("%s/tls.crt", tmpFolder))
	t.Setenv(fmt.Sprintf("%s%s", prefix, envKeyLocationSuffix), fmt.Sprintf("%s/tls.key", tmpFolder))
	t.Setenv(fmt.Sprintf("%s%s", prefix, envCaLocationSuffix), fmt.Sprintf("%s/ca.crt", tmpFolder))
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
	t.Setenv(fmt.Sprintf("%s%s", prefix, envCrtLocationSuffix), fmt.Sprintf("%s/tls.crt", tmpFolder))
	t.Setenv(fmt.Sprintf("%s%s", prefix, envKeyLocationSuffix), fmt.Sprintf("%s/tls.key", tmpFolder))
	t.Setenv(fmt.Sprintf("%s%s", prefix, envCaLocationSuffix), fmt.Sprintf("%s/ca.crt", tmpFolder))
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
}

func TestWatchSecretCertificateWithCa(t *testing.T) {
	g := NewGomegaWithT(t)
	ca, err := os.ReadFile("testdata/cert1/ca.crt")
	g.Expect(err).To(BeNil())
	sec := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "certs", Namespace: "default"},
		Data: map[string][]byte{
			"ca.crt": ca,
		},
	}
	prefix := "SCHEDULER"
	tmpFolder := t.TempDir()
	t.Setenv("SCHEDULER_TLS_SECRET_NAME", sec.Name)
	t.Setenv(fmt.Sprintf("%s%s", prefix, envCaLocationSuffix), fmt.Sprintf("%s/ca.crt", tmpFolder))
	t.Setenv("POD_NAMESPACE", sec.Namespace)
	clientset := fake.NewSimpleClientset(&sec)
	cs, err := NewCertificateStore(Prefix(prefix), ClientSet(clientset), CaOnly(true))
	g.Expect(err).To(BeNil())
	caCreated := cs.GetCertificate().Ca
	g.Expect(caCreated).ToNot(BeNil())
	ca, err = os.ReadFile("testdata/cert2/ca.crt")
	g.Expect(err).To(BeNil())
	sec2 := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "certs", Namespace: "default"},
		Data: map[string][]byte{
			"ca.crt": ca,
		},
	}
	_, err = clientset.CoreV1().Secrets(sec.Namespace).Update(context.Background(), &sec2, metav1.UpdateOptions{})
	g.Expect(err).To(BeNil())
	time.Sleep(time.Millisecond * 500)
	caCreated2 := cs.GetCertificate().Ca
	g.Expect(caCreated2).ToNot(BeNil())
	g.Expect(caCreated).ToNot(Equal(caCreated2))
}
