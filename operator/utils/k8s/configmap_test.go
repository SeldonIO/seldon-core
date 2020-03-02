package k8s

import (
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/fake"
	ctrl "sigs.k8s.io/controller-runtime"
	"testing"
)

func TestConfigmapCreate(t *testing.T) {
	g := NewGomegaWithT(t)
	bytes, err := LoadBytesFromFile("testdata", "configmap.yaml")
	g.Expect(err).To(BeNil())
	client := fake.NewSimpleClientset()

	cc := NewConfigmapCreator(client, ctrl.Log)
	err = cc.CreateConfigmap(bytes, "test")
	g.Expect(err).To(BeNil())
}
