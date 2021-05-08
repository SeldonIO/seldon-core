package k8s

import (
	"context"
	. "github.com/onsi/gomega"
	apiextensionsfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	"k8s.io/apimachinery/pkg/version"
	discovery "k8s.io/client-go/discovery/fake"
	coretesting "k8s.io/client-go/testing"
	ctrl "sigs.k8s.io/controller-runtime"
	"testing"
)

func TestCRDCreateV1beta1(t *testing.T) {
	g := NewGomegaWithT(t)
	apiExtensionsFake := apiextensionsfake.NewSimpleClientset()
	discoveryFake := &discovery.FakeDiscovery{Fake: &coretesting.Fake{}}
	crdCreator := NewCrdCreator(context.TODO(), apiExtensionsFake, discoveryFake, ctrl.Log)
	bytes, err := LoadBytesFromFile("testdata", "crd-v1beta1.yaml")
	g.Expect(err).To(BeNil())
	crd, err := crdCreator.findOrCreateCRDV1beta1(bytes)
	g.Expect(err).To(BeNil())
	g.Expect(crd).NotTo(BeNil())
}

func TestCRDCreateV1(t *testing.T) {
	g := NewGomegaWithT(t)
	apiExtensionsFake := apiextensionsfake.NewSimpleClientset()
	discoveryFake := &discovery.FakeDiscovery{Fake: &coretesting.Fake{},
		FakedServerVersion: &version.Info{
			Major: "1",
			Minor: "18",
		}}
	crdCreator := NewCrdCreator(context.TODO(), apiExtensionsFake, discoveryFake, ctrl.Log)
	bytes, err := LoadBytesFromFile("testdata", "crd-v1.yaml")
	g.Expect(err).To(BeNil())
	crd, err := crdCreator.findOrCreateCRDV1(bytes)
	g.Expect(err).To(BeNil())
	g.Expect(crd).NotTo(BeNil())
}

func TestCRDFindCreateV1(t *testing.T) {
	g := NewGomegaWithT(t)
	apiExtensionsFake := apiextensionsfake.NewSimpleClientset()
	discoveryFake := &discovery.FakeDiscovery{Fake: &coretesting.Fake{},
		FakedServerVersion: &version.Info{
			GitVersion: "v1.19.0",
		}}
	crdCreator := NewCrdCreator(context.TODO(), apiExtensionsFake, discoveryFake, ctrl.Log)
	bytesV1, err := LoadBytesFromFile("testdata", "crd-v1.yaml")
	g.Expect(err).To(BeNil())
	bytesV1beta1, err := LoadBytesFromFile("testdata", "crd-v1beta1.yaml")
	g.Expect(err).To(BeNil())
	crd, err := crdCreator.findOrCreateCRD(bytesV1, bytesV1beta1)
	g.Expect(err).To(BeNil())
	g.Expect(crd).NotTo(BeNil())
}

func TestCRDFindCreateV1Beta1(t *testing.T) {
	g := NewGomegaWithT(t)
	apiExtensionsFake := apiextensionsfake.NewSimpleClientset()
	discoveryFake := &discovery.FakeDiscovery{Fake: &coretesting.Fake{},
		FakedServerVersion: &version.Info{
			GitVersion: "v1.16.0",
		}}
	crdCreator := NewCrdCreator(context.TODO(), apiExtensionsFake, discoveryFake, ctrl.Log)
	bytesV1, err := LoadBytesFromFile("testdata", "crd-v1.yaml")
	g.Expect(err).To(BeNil())
	bytesV1beta1, err := LoadBytesFromFile("testdata", "crd-v1beta1.yaml")
	g.Expect(err).To(BeNil())
	crd, err := crdCreator.findOrCreateCRD(bytesV1, bytesV1beta1)
	g.Expect(err).To(BeNil())
	g.Expect(crd).NotTo(BeNil())
}
