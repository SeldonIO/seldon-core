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
	"testing"

	//+kubebuilder:scaffold:imports
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(
		t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}},
	)
}

var _ = BeforeSuite(func() {
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

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("Controller", func() {
	const (
		Namespace = "default"
	)

	Context("When creating a Pipeline", func() {
		It("Rejects an invalid Spec", func() {
			By("By returning an error")
			ctx := context.Background()

			pipelineName := "-test-pipeline" + strings.Repeat("b", 252)

			pipeline := &mlopsv1alpha1.Pipeline{
				TypeMeta:   metav1.TypeMeta{APIVersion: "batch.tutorial.kubebuilder.io/v1", Kind: "Pipeline"},
				ObjectMeta: metav1.ObjectMeta{Name: pipelineName, Namespace: Namespace},
				Spec:       mlopsv1alpha1.PipelineSpec{},
				Status:     mlopsv1alpha1.PipelineStatus{},
			}

			expectedError := &apierrors.StatusError{
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
				ObjectMeta: metav1.ObjectMeta{Name: "test-pipeline-valid", Namespace: Namespace},
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
				ctx, client.ObjectKey{Name: pipelineName, Namespace: Namespace},
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
				ObjectMeta: metav1.ObjectMeta{Name: "test-model-valid", Namespace: Namespace},
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
				ObjectMeta: metav1.ObjectMeta{Name: "test-model-valid", Namespace: Namespace},
				Spec:       mlopsv1alpha1.ModelSpec{},
				Status:     mlopsv1alpha1.ModelStatus{},
			}

			expectedError := &apierrors.StatusError{
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
				ctx, client.ObjectKey{Name: modelName, Namespace: Namespace},
				retrievedModel,
			)
			Expect(err).To(BeNil())

			Expect(retrievedModel.Spec).To(Equal(expectedModel))
		})
	})
})
