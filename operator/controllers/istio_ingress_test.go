package controllers

import (
	"context"
	"k8s.io/client-go/tools/record"
	"testing"

	. "github.com/onsi/gomega"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	machinelearningv1alpha2 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1alpha2"
	machinelearningv1alpha3 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1alpha3"
	istio_networking "istio.io/api/networking/v1alpha3"
	istio "istio.io/client-go/pkg/apis/networking/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func createScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = v1.AddToScheme(scheme)
	_ = machinelearningv1.AddToScheme(scheme)
	_ = machinelearningv1alpha2.AddToScheme(scheme)
	_ = machinelearningv1alpha3.AddToScheme(scheme)
	_ = v1beta1.AddToScheme(scheme)
	_ = istio.AddToScheme(scheme)
	_ = serializer.NewCodecFactory(scheme).UniversalDeserializer().Decode
	return scheme
}

func createTestSeldonDeploymentForCleaners(mlDepName string, namespace string) *machinelearningv1.SeldonDeployment {
	return &machinelearningv1.SeldonDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mlDepName,
			Namespace: namespace,
		},
		Spec: machinelearningv1.SeldonDeploymentSpec{
			Predictors: []machinelearningv1.PredictorSpec{
				{
					Name: "p1",
					ComponentSpecs: []*machinelearningv1.SeldonPodSpec{
						{
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
					},
				},
			},
		},
	}
}

func createTestVirtualService(instance *machinelearningv1.SeldonDeployment, scheme *runtime.Scheme, name string) (*istio.VirtualService, error) {
	vsvc := &istio.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: instance.Namespace,
		},
		Spec: istio_networking.VirtualService{
			Hosts:    []string{"*"},
			Gateways: []string{"mygateway"},
			Http: []*istio_networking.HTTPRoute{
				{
					Match: []*istio_networking.HTTPMatchRequest{
						{
							Uri: &istio_networking.StringMatch{MatchType: &istio_networking.StringMatch_Prefix{Prefix: "/seldon/" + instance.Namespace + "/" + instance.GetName() + "/"}},
						},
					},
					Rewrite: &istio_networking.HTTPRewrite{Uri: "/"},
				},
			},
		},
	}

	err := ctrl.SetControllerReference(instance, vsvc, scheme)
	return vsvc, err
}

func TestCleanVirtualServices(t *testing.T) {
	g := NewGomegaWithT(t)
	scheme = createScheme()
	client := fake.NewFakeClientWithScheme(scheme)

	mlDepName := "mymodel"
	namespace := "default"
	instance := createTestSeldonDeploymentForCleaners(mlDepName, namespace)

	//Create SeldoNDeployment
	err := client.Create(context.Background(), instance)
	g.Expect(err).To(BeNil())
	foundInstance := &machinelearningv1.SeldonDeployment{}
	err = client.Get(context.Background(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, foundInstance)
	g.Expect(err).To(BeNil())

	// Create Virtual Service
	nameOk := "name-ok"
	vsvcOk, err := createTestVirtualService(foundInstance, scheme, nameOk)
	g.Expect(err).To(BeNil())
	err = client.Create(context.Background(), vsvcOk)
	g.Expect(err).To(BeNil())

	// Virtual Services to be garbage collected
	nameRouge1 := "name-rouge1"
	vsvcRouge1, err := createTestVirtualService(foundInstance, scheme, nameRouge1)
	g.Expect(err).To(BeNil())
	err = client.Create(context.Background(), vsvcRouge1)
	g.Expect(err).To(BeNil())
	nameRouge2 := "name-rouge2"
	vsvcRouge2, err := createTestVirtualService(foundInstance, scheme, nameRouge2)
	g.Expect(err).To(BeNil())
	err = client.Create(context.Background(), vsvcRouge2)
	g.Expect(err).To(BeNil())

	okList := []runtime.Object{vsvcOk}
	logger := ctrl.Log.WithName("controllers").WithName("SeldonDeployment")
	recorder := record.NewFakeRecorder(10)
	ready, deleted, err := cleanupVirtualServices(client, recorder, instance, okList, logger)
	g.Expect(err).To(BeNil())
	g.Expect(ready).To(BeTrue())
	g.Expect(len(deleted)).To(Equal(2))
	g.Expect(deleted[0].Name).To(Equal(nameRouge1))
	g.Expect(deleted[1].Name).To(Equal(nameRouge2))

	// Delete events should be emitted
	expectedEvents := 2
	var events []string
	for i := 0; i < expectedEvents; i++ {
		evt := <-recorder.Events
		events = append(events, evt)
	}
	g.Expect(len(events)).To(Equal(expectedEvents))
	g.Expect(events[0]).To(Equal("Normal DeleteVirtualService Delete VirtualService \"name-rouge1\""))
	g.Expect(events[1]).To(Equal("Normal DeleteVirtualService Delete VirtualService \"name-rouge2\""))
}
