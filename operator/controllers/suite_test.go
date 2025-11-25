/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package controllers

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	v2 "github.com/emissary-ingress/emissary/v3/pkg/api/getambassador.io/v2"
	kedav1alpha1 "github.com/kedacore/keda/v2/apis/keda/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	machinelearningv1alpha2 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1alpha2"
	machinelearningv1alpha3 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1alpha3"
	"github.com/seldonio/seldon-core/operator/constants"
	istio "istio.io/client-go/pkg/apis/networking/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var k8sClient client.Client
var k8sManager ctrl.Manager
var testEnv *envtest.Environment
var scheme = runtime.NewScheme()
var clientset *kubernetes.Clientset

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t,
		"Controller Suite")
}

var configs = map[string]string{
	"predictor_servers": `{
              "TENSORFLOW_SERVER": {
          "protocols" : {
            "tensorflow": {
              "image": "tensorflow/serving",
              "defaultImageVersion": "2.1.0"
              },
            "seldon": {
              "image": "seldonio/tfserving-proxy",
              "defaultImageVersion": "1.3.0-dev"
              }
            }
        },
        "SKLEARN_SERVER": {
          "protocols" : {
            "seldon": {
              "image": "seldonio/sklearnserver",
              "defaultImageVersion": "1.3.0-dev"
              },
            "v2": {
              "image": "seldonio/mlserver",
              "defaultImageVersion": "0.1.0"
              }
            }
        },
        "XGBOOST_SERVER": {
          "protocols" : {
            "seldon": {
              "image": "seldonio/xgboostserver",
              "defaultImageVersion": "1.3.0-dev"
              },
            "v2": {
              "image": "seldonio/mlserver",
              "defaultImageVersion": "0.1.0"
              }
            }
        },
        "MLFLOW_SERVER": {
          "protocols" : {
            "seldon": {
              "image": "seldonio/mlflowserver",
              "defaultImageVersion": "1.3.0-dev"
              },
            "v2": {
              "image": "seldonio/mlserver",
              "defaultImageVersion": "0.1.0"
              }
            }
        },
        "TRITON_SERVER": {
          "protocols" : {
            "v2": {
              "image": "nvcr.io/nvidia/tritonserver",
              "defaultImageVersion": "21.08-py3"
              }
            }
        }
     }`,
	"storageInitializer": `
	{
	"image" : "seldonio/rclone-storage-initializer:1.16.0",
	"memoryRequest": "100Mi",
	"memoryLimit": "1Gi",
	"cpuRequest": "100m",
	"cpuLimit": "1"
	}`,
	"explainer": `
	{
	"image" : "seldonio/alibiexplainer:1.2.0",
	"image_v2" : "seldonio/mlserver:0.6.0"
	}`,
}

const DefaultManagerNamespace = "seldon-system"

// Create configmap
var configMap = &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name:      machinelearningv1.ControllerConfigMapName,
		Namespace: DefaultManagerNamespace,
	},
	Data: configs,
}

var _ = JustBeforeEach(func() {
	envUseExecutor = "true"
	envExecutorImage = "a"
	envExecutorImageRelated = "b"
	envDefaultUser = ""
	envExplainerImage = ""
})

var _ = BeforeSuite(func(done Done) {
	logger := zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true))
	logf.SetLogger(logger)

	By("bootstrapping test environment")

	testEnv = &envtest.Environment{
		AttachControlPlaneOutput: true,
		CRDDirectoryPaths: []string{
			filepath.Join("..", "testing")},
		WebhookInstallOptions: envtest.WebhookInstallOptions{
			Paths:                        []string{"../generated/admissionregistration.k8s.io_v1_validatingwebhookconfiguration_seldon-validating-webhook-configuration.yaml"},
			LocalServingHostExternalName: "localhost",
			LocalServingHost:             "localhost",
		},
	}

	cfg, err := testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())
	cfg.Timeout = time.Second * 10

	clientset, err = kubernetes.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())

	err = os.Setenv(ENV_ISTIO_ENABLED, "true")
	Expect(err).NotTo(HaveOccurred())

	err = os.Setenv(ENV_KEDA_ENABLED, "true")
	Expect(err).NotTo(HaveOccurred())

	err = clientgoscheme.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = apiextensionsv1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = appsv1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = corev1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = machinelearningv1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = istio.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = kedav1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = v2.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = autoscaling.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	k8sManager, err = ctrl.NewManager(cfg, ctrl.Options{
		Scheme:         scheme,
		LeaderElection: false,
		WebhookServer: webhook.NewServer(webhook.Options{
			Host:    testEnv.WebhookInstallOptions.LocalServingHost,
			Port:    testEnv.WebhookInstallOptions.LocalServingPort,
			CertDir: testEnv.WebhookInstallOptions.LocalServingCertDir,
		}),
	})
	Expect(err).ToNot(HaveOccurred())

	k8sClient = k8sManager.GetClient()
	Expect(k8sClient).ToNot(BeNil())

	Expect(k8sClient.Create(context.TODO(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: DefaultManagerNamespace,
		},
	})).NotTo(HaveOccurred())

	err = (&SeldonDeploymentReconciler{
		Client:    k8sManager.GetClient(),
		ClientSet: clientset,
		Log:       ctrl.Log.WithName("controllers").WithName("SeldonDeployment"),
		Scheme:    k8sManager.GetScheme(),
		Recorder:  k8sManager.GetEventRecorderFor(constants.ControllerName),
	}).SetupWithManager(context.Background(), k8sManager, constants.ControllerName)
	Expect(err).ToNot(HaveOccurred())

	Expect(k8sClient.Create(context.TODO(), configMap)).NotTo(HaveOccurred())

	machinelearningv1.C = k8sClient
	machinelearningv1alpha3.C = k8sClient
	machinelearningv1alpha2.C = k8sClient

	err = (&machinelearningv1alpha2.SeldonDeployment{}).SetupWebhookWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	err = (&machinelearningv1alpha3.SeldonDeployment{}).SetupWebhookWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	err = (&machinelearningv1.SeldonDeployment{}).SetupWebhookWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()

		err = k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})
