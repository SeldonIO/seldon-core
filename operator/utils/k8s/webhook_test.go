package k8s

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	machinelearningv1alpha2 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1alpha2"
	machinelearningv1alpha3 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apiextensionsfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	TestNamespace = "test"
)

func int32Ptr(i int32) *int32 { return &i }

func createManagerDeployment(client kubernetes.Interface, namespace string) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: ManagerDeploymentName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "demo",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "demo",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "web",
							Image: "nginx:1.12",
							Ports: []v1.ContainerPort{
								{
									Name:          "http",
									Protocol:      v1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}
	return client.AppsV1().Deployments(namespace).Create(context.TODO(), deployment, v1meta.CreateOptions{})
}

func createTestCRD(client apiextensionsclient.Interface) (*v1beta1.CustomResourceDefinition, error) {
	crd := &v1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: CRDName,
		},
		Spec: v1beta1.CustomResourceDefinitionSpec{
			Group: "machinelearning.seldon.io",
			Scope: v1beta1.ClusterScoped,
			Names: v1beta1.CustomResourceDefinitionNames{
				Plural: "seldondeployments",
				Kind:   "seldondeployment",
			},
			Version: "v1",
		},
	}
	return client.ApiextensionsV1beta1().CustomResourceDefinitions().Create(context.TODO(), crd, v1meta.CreateOptions{})
}

func createScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = v1.AddToScheme(scheme)
	_ = machinelearningv1.AddToScheme(scheme)
	_ = machinelearningv1alpha2.AddToScheme(scheme)
	_ = machinelearningv1alpha3.AddToScheme(scheme)
	_ = v1beta1.AddToScheme(scheme)
	_ = serializer.NewCodecFactory(scheme).UniversalDeserializer().Decode
	return scheme
}

func TestValidatingWebhookCreate(t *testing.T) {
	g := NewGomegaWithT(t)
	scheme := createScheme()
	bytes, err := LoadBytesFromFile("testdata", "validate.yaml")
	g.Expect(err).To(BeNil())
	client := fake.NewSimpleClientset()
	apiExtensionsFake := apiextensionsfake.NewSimpleClientset()
	crd, err := createTestCRD(apiExtensionsFake)
	g.Expect(err).To(BeNil())
	hosts := []string{"seldon-webhook-service.seldon-system", "seldon-webhook-service.seldon-system.svc"}
	certs, err := certSetup(hosts)
	g.Expect(err).To(BeNil())
	wc, err := NewWebhookCreator(client, certs, ctrl.Log, scheme)
	g.Expect(err).To(BeNil())
	err = wc.CreateValidatingWebhookConfigurationFromFile(context.TODO(), bytes, TestNamespace, crd, false)
	g.Expect(err).To(BeNil())
	_, err = client.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Get(context.TODO(), "seldon-validating-webhook", metav1.GetOptions{})
	g.Expect(err).To(BeNil())
}

func TestValidatingWebhookCreateNamespaced(t *testing.T) {
	g := NewGomegaWithT(t)
	scheme := createScheme()
	bytes, err := LoadBytesFromFile("testdata", "validate.yaml")
	g.Expect(err).To(BeNil())
	client := fake.NewSimpleClientset()
	apiExtensionsFake := apiextensionsfake.NewSimpleClientset()
	crd, err := createTestCRD(apiExtensionsFake)
	g.Expect(err).To(BeNil())
	hosts := []string{"seldon-webhook-service.seldon-system", "seldon-webhook-service.seldon-system.svc"}
	certs, err := certSetup(hosts)
	g.Expect(err).To(BeNil())
	wc, err := NewWebhookCreator(client, certs, ctrl.Log, scheme)
	g.Expect(err).To(BeNil())
	err = wc.CreateValidatingWebhookConfigurationFromFile(context.TODO(), bytes, TestNamespace, crd, true)
	g.Expect(err).To(BeNil())
	_, err = client.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Get(context.TODO(), "seldon-validating-webhook-"+TestNamespace, metav1.GetOptions{})
	g.Expect(err).To(BeNil())
}

func TestWebHookSvcCreate(t *testing.T) {
	g := NewGomegaWithT(t)
	scheme := createScheme()
	bytes, err := LoadBytesFromFile("testdata", "svc.yaml")
	g.Expect(err).To(BeNil())
	client := fake.NewSimpleClientset()
	dep, err := createManagerDeployment(client, TestNamespace)
	g.Expect(err).To(BeNil())
	hosts := []string{"seldon-webhook-service.seldon-system", "seldon-webhook-service.seldon-system.svc"}
	certs, err := certSetup(hosts)
	g.Expect(err).To(BeNil())
	wc, err := NewWebhookCreator(client, certs, ctrl.Log, scheme)
	g.Expect(err).To(BeNil())
	err = wc.CreateWebhookServiceFromFile(context.TODO(), bytes, TestNamespace, dep)
	g.Expect(err).To(BeNil())
}
