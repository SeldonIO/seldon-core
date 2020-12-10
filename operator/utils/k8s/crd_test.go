package k8s

import (
	"context"
	. "github.com/onsi/gomega"
	apiextensionsfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	ctrl "sigs.k8s.io/controller-runtime"
	"testing"
)

func TestCRDCreate(t *testing.T) {
	g := NewGomegaWithT(t)
	apiExtensionsFake := apiextensionsfake.NewSimpleClientset()
	crdCreator := NewCrdCreator(context.TODO(), apiExtensionsFake, ctrl.Log)
	bytes, err := LoadBytesFromFile("testdata", "crd.yaml")
	g.Expect(err).To(BeNil())
	crd, err := crdCreator.findOrCreateCRD(bytes)
	g.Expect(err).To(BeNil())
	g.Expect(crd).NotTo(BeNil())
}
