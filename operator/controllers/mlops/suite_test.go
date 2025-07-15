/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package mlops

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	//+kubebuilder:scaffold:imports
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/scheduler/mock"
	"go.uber.org/mock/gomock"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	k8sClient client.Client
	testEnv   *envtest.Environment
	ctx       context.Context
	cancel    context.CancelFunc

	mockCtrl *gomock.Controller
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(
		t,
		"Controller Suite",
		[]Reporter{},
	)
}

const (
	TestNamespace = "default"
)

type expectedCallInfo struct {
	spec    mlopsv1alpha1.ScalingSpec
	success *atomic.Bool
}

var expectedServerNotifyCalls = make(map[string]expectedCallInfo)

var _ = BeforeSuite(func() {
	ctx, cancel = context.WithCancel(context.TODO())

	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = mlopsv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	mockCtrl = gomock.NewController(GinkgoT())
	schedulerMock := mock.NewMockClient(mockCtrl)

	// due to fact that ServerNotify on scheduler will be called asynchronously, if we just setup
	// the mock in the standard way of expecting a call signature we get failures of missed calls as the
	// test finishes before call is made or failures of wrong arguments used in call. Instead when expected
	// call is received, expectedServerNotifyCalls.success is marked as true and the test case verifies it's true with:
	// 			Eventually(func() bool {
	//				return serverNotifyCalled.Load()
	//			}, "2s", "10ms").Should(BeTrue())
	schedulerMock.EXPECT().ServerNotify(gomock.Any(), gomock.Any(), gomock.Any(), false).
		DoAndReturn(func(_ context.Context, _ *scheduler.SchedulerClient, servers []mlopsv1alpha1.Server, isFirstSync bool) error {

			for _, server := range servers {
				for serverName, check := range expectedServerNotifyCalls {
					if serverName != server.Name {
						continue
					}
					Expect(server.Spec.ScalingSpec).To(Equal(check.spec))
					check.success.Store(true)
				}
			}

			return nil
		}).AnyTimes()

	k8sClient = k8sManager.GetClient()

	serverReconciler := &ServerReconciler{
		Client:    k8sManager.GetClient(),
		Scheme:    k8sManager.GetScheme(),
		Scheduler: schedulerMock,
	}

	err = serverReconciler.SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	err = (&ServerConfigReconciler{
		Client: k8sManager.GetClient(),
		Scheme: k8sManager.GetScheme(),
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()

}, 60)

func addExpectedServerNotifyCall(serverName string, scalingSpec mlopsv1alpha1.ScalingSpec, successNotify *atomic.Bool) {
	expectedServerNotifyCalls[serverName] = expectedCallInfo{
		spec:    scalingSpec,
		success: successNotify,
	}
}

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	cancel()
	mockCtrl.Finish()
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("Controller", func() {

	Context("When creating a Pipeline", func() {
		It("Rejects an invalid Spec", func() {
			By("By returning an error")
			ctx := context.Background()

			pipelineName := "-test-pipeline" + strings.Repeat("b", 252)

			pipeline := &mlopsv1alpha1.Pipeline{
				TypeMeta:   metav1.TypeMeta{APIVersion: "batch.tutorial.kubebuilder.io/v1", Kind: "Pipeline"},
				ObjectMeta: metav1.ObjectMeta{Name: pipelineName, Namespace: TestNamespace},
				Spec:       mlopsv1alpha1.PipelineSpec{},
				Status:     mlopsv1alpha1.PipelineStatus{},
			}

			expectedError := &errors.StatusError{
				ErrStatus: metav1.Status{
					TypeMeta: metav1.TypeMeta{Kind: "", APIVersion: ""},
					ListMeta: metav1.ListMeta{
						SelfLink:           "",
						ResourceVersion:    "",
						Continue:           "",
						RemainingItemCount: nil,
					},
					Status:  "Failure",
					Message: fmt.Sprintf("Pipeline.mlops.seldon.io \"%s\" is invalid: [metadata.name: Invalid value: \"%s\": must be no more than 253 characters, metadata.name: Invalid value: \"%s\": a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*'), spec.steps: Required value]", pipelineName, pipelineName, pipelineName),
					Reason:  metav1.StatusReasonInvalid,
					Details: &metav1.StatusDetails{
						Name:  pipelineName,
						Group: "mlops.seldon.io",
						Kind:  "Pipeline",
						UID:   "",
						Causes: []metav1.StatusCause{
							{
								Type:    metav1.CauseTypeFieldValueInvalid,
								Message: fmt.Sprintf("Invalid value: \"%s\": must be no more than 253 characters", pipelineName),
								Field:   "metadata.name",
							},
							{
								Type:    metav1.CauseTypeFieldValueInvalid,
								Message: fmt.Sprintf("Invalid value: \"%s\": a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')", pipelineName),
								Field:   "metadata.name",
							},
							{
								Type:    metav1.CauseTypeFieldValueRequired,
								Message: "Required value",
								Field:   "spec.steps",
							},
						},
						RetryAfterSeconds: 0,
					},
					Code: http.StatusUnprocessableEntity,
				},
			}

			Expect(k8sClient.Create(ctx, pipeline)).Should(MatchError(expectedError))
		})

		It("Accepts a valid Spec", func() {
			By("By returning nil")
			ctx := context.Background()

			pipeline := &mlopsv1alpha1.Pipeline{
				TypeMeta:   metav1.TypeMeta{APIVersion: "batch.tutorial.kubebuilder.io/v1", Kind: "Pipeline"},
				ObjectMeta: metav1.ObjectMeta{Name: "test-pipeline-valid", Namespace: TestNamespace},
				Spec: mlopsv1alpha1.PipelineSpec{
					Steps: []mlopsv1alpha1.PipelineStep{
						{
							Name: "step-1",
						},
					},
				},
				Status: mlopsv1alpha1.PipelineStatus{},
			}

			Expect(k8sClient.Create(ctx, pipeline)).Should(Succeed())
		})

		It("Retrieves a pipeline by name", func() {
			By("By fetching the pipeline by name")
			ctx := context.Background()
			pipelineName := "test-pipeline-valid"

			retrievedPipeline := &mlopsv1alpha1.Pipeline{}

			// Default value, as per kubebuilder annotation
			expectedInputsJoinType := mlopsv1alpha1.JoinTypeInner

			expectedPipeline :=
				mlopsv1alpha1.PipelineSpec{
					Steps: []mlopsv1alpha1.PipelineStep{
						{
							Name:             "step-1",
							Inputs:           nil,
							JoinWindowMs:     nil,
							TensorMap:        nil,
							Triggers:         nil,
							InputsJoinType:   &expectedInputsJoinType,
							TriggersJoinType: nil,
							Batch:            nil,
						}},
				}

			err := k8sClient.Get(
				ctx, client.ObjectKey{Name: pipelineName, Namespace: TestNamespace},
				retrievedPipeline,
			)
			Expect(err).To(BeNil())

			Expect(retrievedPipeline.Spec).To(Equal(expectedPipeline))
		})
	})

	Context("When creating a Model", func() {
		It("Accepts a valid Spec", func() {
			By("By returning nil")
			ctx := context.Background()

			model := &mlopsv1alpha1.Model{
				TypeMeta:   metav1.TypeMeta{APIVersion: "batch.tutorial.kubebuilder.io/v1", Kind: "Model"},
				ObjectMeta: metav1.ObjectMeta{Name: "test-model-valid", Namespace: TestNamespace},
				Spec:       mlopsv1alpha1.ModelSpec{},
				Status:     mlopsv1alpha1.ModelStatus{},
			}

			Expect(k8sClient.Create(ctx, model)).Should(Succeed())
		})

		// This relies on the previous test having run
		It("Rejects a Spec with the name of a model that already exists", func() {
			By("By returning an error")
			ctx := context.Background()
			model := &mlopsv1alpha1.Model{
				TypeMeta:   metav1.TypeMeta{APIVersion: "batch.tutorial.kubebuilder.io/v1", Kind: "Model"},
				ObjectMeta: metav1.ObjectMeta{Name: "test-model-valid", Namespace: TestNamespace},
				Spec:       mlopsv1alpha1.ModelSpec{},
				Status:     mlopsv1alpha1.ModelStatus{},
			}

			expectedError := &errors.StatusError{
				ErrStatus: metav1.Status{
					TypeMeta: metav1.TypeMeta{Kind: "", APIVersion: ""},
					ListMeta: metav1.ListMeta{
						SelfLink:           "",
						ResourceVersion:    "",
						Continue:           "",
						RemainingItemCount: nil,
					},
					Status:  "Failure",
					Message: "models.mlops.seldon.io \"test-model-valid\" already exists",
					Reason:  metav1.StatusReasonAlreadyExists,
					Details: &metav1.StatusDetails{
						Name:              "test-model-valid",
						Group:             "mlops.seldon.io",
						Kind:              "models",
						UID:               "",
						Causes:            nil,
						RetryAfterSeconds: 0,
					},
					Code: http.StatusConflict,
				},
			}

			Expect(k8sClient.Create(ctx, model)).Should(MatchError(expectedError))
		})

		It("Retrieves a model by name", func() {
			By("By fetching the model by name")
			ctx := context.Background()
			modelName := "test-model-valid" // Replace with the actual name you want to retrieve

			retrievedModel := &mlopsv1alpha1.Model{}

			expectedModel :=
				mlopsv1alpha1.ModelSpec{
					ScalingSpec: mlopsv1alpha1.ScalingSpec{
						Replicas:    nil,
						MinReplicas: nil,
						MaxReplicas: nil,
					},
				}

			// Fetch the model by name
			err := k8sClient.Get(
				ctx, client.ObjectKey{Name: modelName, Namespace: TestNamespace},
				retrievedModel,
			)
			Expect(err).To(BeNil())

			Expect(retrievedModel.Spec).To(Equal(expectedModel))
		})
	})

	Context("When creating a ServerConfig spec", func() {
		It("Accepts a valid Spec", func() {
			ctx := context.Background()

			serverConfig := &mlopsv1alpha1.ServerConfig{
				TypeMeta:   metav1.TypeMeta{APIVersion: "batch.tutorial.kubebuilder.io/v1", Kind: "ServerConfig"},
				ObjectMeta: metav1.ObjectMeta{Name: "mlserver", Namespace: TestNamespace},
				Spec: mlopsv1alpha1.ServerConfigSpec{
					VolumeClaimTemplates: []mlopsv1alpha1.PersistentVolumeClaim{
						{
							Name: "vol1",
							Spec: corev1.PersistentVolumeClaimSpec{},
						},
					},
					PodSpec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "container-1",
								Image: "mlserver",
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "volume-1",
										MountPath: "/mnt",
									},
								},
							},
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, serverConfig)).Should(Succeed())

			Eventually(func(g Gomega) {
				By("By fetching the ServerConfig by name")
				ctx := context.Background()
				serverConfigName := "mlserver"

				serverConfig := &mlopsv1alpha1.ServerConfig{}
				err := k8sClient.Get(
					ctx, client.ObjectKey{Name: serverConfigName, Namespace: TestNamespace},
					serverConfig,
				)
				g.Expect(err).To(BeNil())
				g.Expect(serverConfig.Name).To(Equal(serverConfigName))
			}).Should(Succeed())
		})
	})

	Context("When creating a Server spec", func() {
		It("Accepts a valid Spec with minReplicas = 1 replicas = 1 ", func() {
			testID := "test-mlserver"

			serverNotifyCalled := &atomic.Bool{}
			addExpectedServerNotifyCall(testID, mlopsv1alpha1.ScalingSpec{
				MinReplicas: ptr.To(int32(1)),
				Replicas:    ptr.To(int32(1)),
				MaxReplicas: ptr.To(int32(2)),
			}, serverNotifyCalled)

			ctx := context.Background()

			server := &mlopsv1alpha1.Server{
				TypeMeta:   metav1.TypeMeta{APIVersion: "batch.tutorial.kubebuilder.io/v1", Kind: "Server"},
				ObjectMeta: metav1.ObjectMeta{Name: testID, Namespace: TestNamespace},
				Spec: mlopsv1alpha1.ServerSpec{
					ServerConfig: "mlserver",
					StatefulSetPersistentVolumeClaimRetentionPolicy: nil,
					ScalingSpec: mlopsv1alpha1.ScalingSpec{
						MinReplicas: ptr.To(int32(1)),
						Replicas:    ptr.To(int32(1)),
						MaxReplicas: ptr.To(int32(2)),
					},
					DisableAutoUpdate: false,
				},
			}

			Expect(k8sClient.Create(ctx, server)).Should(Succeed())

			// we need to wait for the expected call to ServerNotify
			Eventually(func() bool {
				return serverNotifyCalled.Load()
			}, "2s", "10ms").Should(BeTrue())

			Eventually(func(g Gomega) {
				By("By fetching the StatefulSet by name")
				ctx := context.Background()

				statefulset := &v1.StatefulSet{}
				err := k8sClient.Get(
					ctx, client.ObjectKey{Name: testID, Namespace: TestNamespace},
					statefulset,
				)

				g.Expect(err).To(BeNil())
				g.Expect(statefulset.Name).To(Equal(testID))
				g.Expect(statefulset.Spec.Replicas).To(Equal(ptr.To(int32(1))))
			}).Should(Succeed())
		})
	})

	Context("When creating a Server spec", func() {
		It("Accepts a valid Spec with minReplicas = 2 replicas not set, therefore replicas should take minReplicas value of 1", func() {
			testID := "test-mlserver-2"
			minReplicas := ptr.To(int32(2))

			serverNotifyCalled := &atomic.Bool{}
			addExpectedServerNotifyCall(testID, mlopsv1alpha1.ScalingSpec{
				MinReplicas: minReplicas,
				MaxReplicas: ptr.To(int32(2)),
			}, serverNotifyCalled)

			ctx := context.Background()

			server := &mlopsv1alpha1.Server{
				TypeMeta:   metav1.TypeMeta{APIVersion: "batch.tutorial.kubebuilder.io/v1", Kind: "Server"},
				ObjectMeta: metav1.ObjectMeta{Name: testID, Namespace: TestNamespace},
				Spec: mlopsv1alpha1.ServerSpec{
					ServerConfig: "mlserver",
					ScalingSpec: mlopsv1alpha1.ScalingSpec{
						MinReplicas: minReplicas,
						MaxReplicas: ptr.To(int32(2)),
					},
				},
			}

			Expect(k8sClient.Create(ctx, server)).Should(Succeed())

			// we need to wait for the expected call to ServerNotify
			Eventually(func() bool {
				return serverNotifyCalled.Load()
			}, "2s", "10ms").Should(BeTrue())

			Eventually(func(g Gomega) {
				By("By fetching the StatefulSet by name")
				ctx := context.Background()

				statefulset := &v1.StatefulSet{}
				err := k8sClient.Get(
					ctx, client.ObjectKey{Name: testID, Namespace: TestNamespace},
					statefulset,
				)

				g.Expect(err).To(BeNil())
				g.Expect(statefulset.Name).To(Equal(testID))
				g.Expect(statefulset.Spec.Replicas).To(Equal(minReplicas))
			}).WithTimeout(5 * time.Second).Should(Succeed())
		})
	})

	Context("When creating a Server spec", func() {
		It("Rejects an invalid Spec with minReplicas = 2 replicas = 1", func() {
			testID := "test-mlserver-3"
			minReplicas := ptr.To(int32(2))
			replicas := ptr.To(int32(1))

			serverNotifyCalled := &atomic.Bool{}
			addExpectedServerNotifyCall(testID, mlopsv1alpha1.ScalingSpec{
				MinReplicas: minReplicas,
				Replicas:    replicas,
			}, serverNotifyCalled)

			ctx := context.Background()

			server := &mlopsv1alpha1.Server{
				TypeMeta:   metav1.TypeMeta{APIVersion: "batch.tutorial.kubebuilder.io/v1", Kind: "Server"},
				ObjectMeta: metav1.ObjectMeta{Name: testID, Namespace: TestNamespace},
				Spec: mlopsv1alpha1.ServerSpec{
					ServerConfig: "mlserver",
					ScalingSpec: mlopsv1alpha1.ScalingSpec{
						MinReplicas: minReplicas,
						Replicas:    replicas,
					},
				},
			}

			Expect(k8sClient.Create(ctx, server)).Should(Succeed())

			// we expect ServerNotify not to be called
			Eventually(func() bool {
				return serverNotifyCalled.Load()
			}, "2s", "10ms").Should(BeFalse())

			Eventually(func(g Gomega) {
				By("By fetching the StatefulSet by name, verify does not exist")
				ctx := context.Background()

				statefulset := &v1.StatefulSet{}
				err := k8sClient.Get(
					ctx, client.ObjectKey{Name: testID, Namespace: TestNamespace},
					statefulset,
				)

				g.Expect(err).ShouldNot(BeNil())
				errStatus, ok := err.(*errors.StatusError)
				g.Expect(ok).To(BeTrue())
				g.Expect(errors.IsNotFound(errStatus)).Should(BeTrue())
			}).WithTimeout(5 * time.Second).Should(Succeed())
		})
	})
})
