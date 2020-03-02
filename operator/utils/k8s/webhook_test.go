package k8s

import (
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/fake"
	ctrl "sigs.k8s.io/controller-runtime"
	"testing"
)

func TestMutatingWebhookCreate(t *testing.T) {
	g := NewGomegaWithT(t)
	bytes, err := LoadBytesFromFile("testdata", "mutate.yaml")
	g.Expect(err).To(BeNil())
	client := fake.NewSimpleClientset()
	hosts := []string{"seldon-webhook-service.seldon-system", "seldon-webhook-service.seldon-system.svc"}
	certs, err := certSetup(hosts)
	g.Expect(err).To(BeNil())
	wc, err := NewWebhookCreator(client, certs, ctrl.Log)
	g.Expect(err).To(BeNil())
	err = wc.CreateMutatingWebhookConfigurationFromFile(bytes)
	g.Expect(err).To(BeNil())
}

func TestWebHookSvcCreate(t *testing.T) {
	g := NewGomegaWithT(t)
	bytes, err := LoadBytesFromFile("testdata", "svc.yaml")
	g.Expect(err).To(BeNil())
	client := fake.NewSimpleClientset()
	hosts := []string{"seldon-webhook-service.seldon-system", "seldon-webhook-service.seldon-system.svc"}
	certs, err := certSetup(hosts)
	g.Expect(err).To(BeNil())
	wc, err := NewWebhookCreator(client, certs, ctrl.Log)
	g.Expect(err).To(BeNil())
	err = wc.CreateWebhookServiceFromFile(bytes, "test")
	g.Expect(err).To(BeNil())
}
