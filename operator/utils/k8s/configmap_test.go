package k8s

import (
	"context"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/fake"
	ctrl "sigs.k8s.io/controller-runtime"
	"testing"
)

func TestConfigmapCreate(t *testing.T) {
	g := NewGomegaWithT(t)
	scheme := createScheme()
	bytes, err := LoadBytesFromFile("testdata", "configmap.yaml")
	g.Expect(err).To(BeNil())
	client := fake.NewSimpleClientset()
	dep, err := createManagerDeployment(client, TestNamespace)
	g.Expect(err).To(BeNil())
	cc := NewConfigmapCreator(client, ctrl.Log, scheme)
	err = cc.CreateConfigmap(context.TODO(), bytes, TestNamespace, dep)
	g.Expect(err).To(BeNil())
}
