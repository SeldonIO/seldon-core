package k8s

import (
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"testing"
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
	return client.AppsV1().Deployments(namespace).Create(deployment)
}

func createScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = v1.AddToScheme(scheme)
	return scheme
}

func TestMutatingWebhookCreate(t *testing.T) {
	g := NewGomegaWithT(t)
	scheme := createScheme()
	bytes, err := LoadBytesFromFile("testdata", "mutate.yaml")
	g.Expect(err).To(BeNil())
	client := fake.NewSimpleClientset()
	dep, err := createManagerDeployment(client, TestNamespace)
	g.Expect(err).To(BeNil())
	hosts := []string{"seldon-webhook-service.seldon-system", "seldon-webhook-service.seldon-system.svc"}
	certs, err := certSetup(hosts)
	g.Expect(err).To(BeNil())
	wc, err := NewWebhookCreator(client, certs, ctrl.Log, scheme)
	g.Expect(err).To(BeNil())
	err = wc.CreateMutatingWebhookConfigurationFromFile(bytes, TestNamespace, dep, false)
	g.Expect(err).To(BeNil())
}

func TestValidatingWebhookCreate(t *testing.T) {
	g := NewGomegaWithT(t)
	scheme := createScheme()
	bytes, err := LoadBytesFromFile("testdata", "mutate.yaml")
	g.Expect(err).To(BeNil())
	client := fake.NewSimpleClientset()
	dep, err := createManagerDeployment(client, TestNamespace)
	g.Expect(err).To(BeNil())
	hosts := []string{"seldon-webhook-service.seldon-system", "seldon-webhook-service.seldon-system.svc"}
	certs, err := certSetup(hosts)
	g.Expect(err).To(BeNil())
	wc, err := NewWebhookCreator(client, certs, ctrl.Log, scheme)
	g.Expect(err).To(BeNil())
	err = wc.CreateValidatingWebhookConfigurationFromFile(bytes, TestNamespace, dep, false)
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
	err = wc.CreateWebhookServiceFromFile(bytes, TestNamespace, dep)
	g.Expect(err).To(BeNil())
}
