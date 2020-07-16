package controllers

import (
	. "github.com/onsi/gomega"
	contour "github.com/projectcontour/contour/apis/projectcontour/v1"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"gopkg.in/yaml.v2"
	"io"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"strings"
	"testing"
)

func createSeldonDeploymentForIngressTest(name string, namespace string) *machinelearningv1.SeldonDeployment {
	envEngineImage = "seldonio/engine:0.1"
	modelType := machinelearningv1.MODEL
	instance := &machinelearningv1.SeldonDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: machinelearningv1.SeldonDeploymentSpec{
			Predictors: []machinelearningv1.PredictorSpec{
				{
					Name: "p1",
					ComponentSpecs: []*machinelearningv1.SeldonPodSpec{
						{
							Metadata: metav1.ObjectMeta{},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "seldonio/mock_classifier:1.0",
										Name:  "classifier",
									},
								},
							},
						},
					},
					Graph: machinelearningv1.PredictiveUnit{
						Name: "classifier",
						Type: &modelType,
					},
				},
			},
		},
	}
	return instance
}

// TODO(jpg) make environment variables overridable from tests so can test multi-vhost as well
func TestContourIngressSingleVirtualHost(t *testing.T) {
	g := NewGomegaWithT(t)
	name := "dep"
	namespace := "default"

	logger := ctrl.Log.WithName("controllers").WithName("SeldonDeployment")
	reconciler := &SeldonDeploymentReconciler{
		Log:       logger,
		Ingresses: []Ingress{NewContourIngress()},
	}

	instance := createSeldonDeploymentForIngressTest(name, namespace)
	instance.Spec.DefaultSeldonDeployment(name, namespace)
	c, err := reconciler.createComponents(instance, nil, logger)
	g.Expect(err).To(BeNil())

	// We should have only created Contour HTTPProxies
	g.Expect(len(c.ingressResources)).To(Equal(1))

	// Check HTTPProxy resources are created correctly
	httpProxies, ok := c.ingressResources[ContourHTTPProxies]
	g.Expect(ok).To(BeTrue())
	g.Expect(len(httpProxies)).To(Equal(1))
	httpProxy := httpProxies[0].(*contour.HTTPProxy)
	g.Expect(httpProxy.Name).To(Equal(name))
	g.Expect(httpProxy.Spec.VirtualHost).To(BeNil())
}

func TestIstioIngress(t *testing.T) {
	g := NewGomegaWithT(t)
	name := "dep"
	namespace := "default"

	logger := ctrl.Log.WithName("controllers").WithName("SeldonDeployment")
	reconciler := &SeldonDeploymentReconciler{
		Log:       logger,
		Ingresses: []Ingress{NewIstioIngress()},
	}

	instance := createSeldonDeploymentForIngressTest(name, namespace)
	instance.Spec.DefaultSeldonDeployment(name, namespace)
	c, err := reconciler.createComponents(instance, nil, logger)
	g.Expect(err).To(BeNil())

	// Both VirtualServices and DestinationRules are created when using Istio ingress
	g.Expect(len(c.ingressResources)).To(Equal(2))

	// Check VirtualService resources are created correctly
	vsvcs, ok := c.ingressResources[IstioVirtualServices]
	g.Expect(ok).To(BeTrue())
	// There should be 2 vsvcs created, one for each protocol
	g.Expect(len(vsvcs)).To(Equal(2))

	// Check VirtualService resources are created correctly
	drules, ok := c.ingressResources[IstioDestinationRules]
	g.Expect(ok).To(BeTrue())
	g.Expect(len(drules)).To(Equal(1))
}

func TestMultipleIngress(t *testing.T) {
	g := NewGomegaWithT(t)
	name := "dep"
	namespace := "default"

	logger := ctrl.Log.WithName("controllers").WithName("SeldonDeployment")
	reconciler := &SeldonDeploymentReconciler{
		Log:       logger,
		Ingresses: []Ingress{NewIstioIngress(), NewContourIngress()},
	}

	instance := createSeldonDeploymentForIngressTest(name, namespace)
	instance.Spec.DefaultSeldonDeployment(name, namespace)
	c, err := reconciler.createComponents(instance, nil, logger)
	g.Expect(err).To(BeNil())

	// Expect VirtualService, DestinationRule and HTTPProxy resources
	g.Expect(len(c.ingressResources)).To(Equal(3))

	// Check VirtualService resources are created correctly
	vsvcs, ok := c.ingressResources[IstioVirtualServices]
	g.Expect(ok).To(BeTrue())
	// There should be 2 vsvcs created, one for each protocol
	g.Expect(len(vsvcs)).To(Equal(2))

	// Check VirtualService resources are created correctly
	drules, ok := c.ingressResources[IstioDestinationRules]
	g.Expect(ok).To(BeTrue())
	g.Expect(len(drules)).To(Equal(1))

	// Check HTTPProxy resources are created correctly
	httpProxies, ok := c.ingressResources[ContourHTTPProxies]
	g.Expect(ok).To(BeTrue())
	g.Expect(len(httpProxies)).To(Equal(1))
}

func TestAmbassadorIngress(t *testing.T) {
	g := NewGomegaWithT(t)
	name := "dep"
	namespace := "default"

	logger := ctrl.Log.WithName("controllers").WithName("SeldonDeployment")
	reconciler := &SeldonDeploymentReconciler{
		Log:       logger,
		Ingresses: []Ingress{NewAmbassadorIngress()},
	}

	instance := createSeldonDeploymentForIngressTest(name, namespace)
	instance.Spec.DefaultSeldonDeployment(name, namespace)
	c, err := reconciler.createComponents(instance, nil, logger)
	g.Expect(err).To(BeNil())

	// Ambassador doesn't create extra resources
	g.Expect(len(c.ingressResources)).To(Equal(0))

	// Check Ambassador annotation is created on Service
	g.Expect(len(c.services)).To(Equal(2))
	pSvc := c.services[1]
	g.Expect(pSvc.Name).To(Equal("dep-p1"))
	g.Expect(len(pSvc.Annotations)).To(Equal(1))
	annotation, ok := pSvc.Annotations[AMBASSADOR_ANNOTATION]
	g.Expect(ok).To(BeTrue())
	reader := strings.NewReader(annotation)
	decoder := yaml.NewDecoder(reader)
	var mappings []*AmbassadorConfig
	for {
		var mapping AmbassadorConfig
		err = decoder.Decode(&mapping)
		if err == io.EOF {
			break
		}
		g.Expect(err).To(BeNil())
		mappings = append(mappings, &mapping)
	}
	// We expect a mapping for each protocol
	g.Expect(len(mappings)).To(Equal(2))
	g.Expect(err).To(BeNil())
}
