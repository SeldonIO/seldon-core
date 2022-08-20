package tls

import (
	"context"
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
	key := "SCHEDULER_TLS_FOLDER_PATH"
	t.Setenv(key, tmpFolder)
	clientset := fake.NewSimpleClientset()
	cs, err := NewCertificateStore(prefix, clientset)
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
	key := "SCHEDULER_TLS_SECRET_NAME"
	t.Setenv(key, sec.Name)
	t.Setenv("POD_NAMESPACE", sec.Namespace)
	clientset := fake.NewSimpleClientset(&sec)
	cs, err := NewCertificateStore(prefix, clientset)
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
