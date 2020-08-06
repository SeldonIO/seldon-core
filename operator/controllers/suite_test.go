/*
Copyright 2019 The Seldon Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"github.com/seldonio/seldon-core/operator/constants"
	istio "istio.io/client-go/pkg/apis/networking/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"os"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"testing"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var k8sManager ctrl.Manager
var testEnv *envtest.Environment
var scheme = runtime.NewScheme()
var clientset *kubernetes.Clientset

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var configs = map[string]string{
	"predictor_servers": `{
             "TENSORFLOW_SERVER": {
                 "tensorflow": true,
                 "tfImage": "tensorflow/serving:2.1",
                 "rest": {
                   "image": "seldonio/tfserving-proxy_rest",
                   "defaultImageVersion": "0.7"
                 },
                 "grpc": {
                   "image": "seldonio/tfserving-proxy_grpc",
                   "defaultImageVersion": "0.7"
                 }
             },
             "SKLEARN_SERVER": {
                 "rest": {
                   "image": "seldonio/sklearnserver_rest",
                   "defaultImageVersion": "0.2"
                 },
                 "grpc": {
                   "image": "seldonio/sklearnserver_grpc",
                   "defaultImageVersion": "0.2"
                 }
             },
             "XGBOOST_SERVER": {
                 "rest": {
                   "image": "seldonio/xgboostserver_rest",
                   "defaultImageVersion": "0.2"
                 },
                 "grpc": {
                   "image": "seldonio/xgboostserver_grpc",
                   "defaultImageVersion": "0.2"
                 }
             },
             "MLFLOW_SERVER": {
                 "rest": {
                   "image": "seldonio/mlflowserver_rest",
                   "defaultImageVersion": "0.2"
                 },
                 "grpc": {
                   "image": "seldonio/mlflowserver_grpc",
                   "defaultImageVersion": "0.2"
                 }
             }
         }`,
	"storageInitializer": `
	{
	"image" : "gcr.io/kfserving/storage-initializer:0.2.2",
	"memoryRequest": "100Mi",
	"memoryLimit": "1Gi",
	"cpuRequest": "100m",
	"cpuLimit": "1"
	}`,
	"explainer": `
	{
	"image" : "seldonio/alibiexplainer:1.2.0"
	}`,
}

// Create configmap
var configMap = &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name:      machinelearningv1.ControllerConfigMapName,
		Namespace: "seldon-system",
	},
	Data: configs,
}

var _ = JustBeforeEach(func() {
	envUseExecutor = "true"
	envExecutorImage = "a"
	envExecutorImageRelated = "b"
	envEngineImage = "c"
	envEngineImageRelated = "d"
	envDefaultUser = ""
	envExplainerImage = ""
})

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	By("bootstrapping test environment")

	//apiServerFlags := envtest.DefaultKubeAPIServerFlags[0 : len(envtest.DefaultKubeAPIServerFlags)-1]
	//apiServerFlags = append(apiServerFlags, "--admission-control=MutatingAdmissionWebhook")

	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			//filepath.Join("..", "config", "crd", "bases"),
			filepath.Join("..", "testing")},
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	clientset, err = kubernetes.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())

	err = os.Setenv(ENV_ISTIO_ENABLED, "true")
	Expect(err).NotTo(HaveOccurred())

	err = clientgoscheme.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	//err = scheme.AddToScheme(scheme.Scheme)
	//Expect(err).NotTo(HaveOccurred())

	err = appsv1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = corev1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = machinelearningv1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = istio.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	k8sManager, err = ctrl.NewManager(cfg, ctrl.Options{
		Scheme:         scheme,
		LeaderElection: false,
	})
	Expect(err).ToNot(HaveOccurred())

	err = (&SeldonDeploymentReconciler{
		Client:    k8sManager.GetClient(),
		ClientSet: clientset,
		Log:       ctrl.Log.WithName("controllers").WithName("SeldonDeployment"),
		Scheme:    k8sManager.GetScheme(),
		Recorder:  k8sManager.GetEventRecorderFor(constants.ControllerName),
	}).SetupWithManager(k8sManager, constants.ControllerName)
	Expect(err).ToNot(HaveOccurred())

	//k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	k8sClient = k8sManager.GetClient()
	Expect(k8sClient).ToNot(BeNil())

	Expect(k8sClient.Create(context.TODO(), configMap)).NotTo(HaveOccurred())
	//	defer k8sClient.Delete(context.TODO(), configMap)

	machinelearningv1.C = k8sClient

	fmt.Println("test k8s client")
	fmt.Printf("%+v\n", k8sClient)

	go func() {
		defer GinkgoRecover()

		//can't call webhook as leads to https://github.com/kubernetes-sigs/controller-runtime/issues/491
		//err = (&machinelearningv1.SeldonDeployment{}).SetupWebhookWithManager(k8sManager)
		//Expect(err).ToNot(HaveOccurred())
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
