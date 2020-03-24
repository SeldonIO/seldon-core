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

package main

import (
	"flag"
	"github.com/seldonio/seldon-core/operator/constants"
	"os"

	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning/v1"
	machinelearningv1alpha2 "github.com/seldonio/seldon-core/operator/apis/machinelearning/v1alpha2"
	machinelearningv1alpha3 "github.com/seldonio/seldon-core/operator/apis/machinelearning/v1alpha3"
	"github.com/seldonio/seldon-core/operator/controllers"
	k8sutils "github.com/seldonio/seldon-core/operator/utils/k8s"
	istio "istio.io/client-go/pkg/apis/networking/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = machinelearningv1.AddToScheme(scheme)
	_ = machinelearningv1alpha2.AddToScheme(scheme)
	_ = machinelearningv1alpha3.AddToScheme(scheme)
	_ = v1beta1.AddToScheme(scheme)
	if controllers.GetEnv(controllers.ENV_ISTIO_ENABLED, "false") == "true" {
		_ = istio.AddToScheme(scheme)
	}
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var webHookPort int
	var namespace string
	var operatorNamespace string
	var createResources bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.IntVar(&webHookPort, "webhook-port", 443, "Webhook server port")
	flag.StringVar(&namespace, "namespace", "", "The namespace to restrict the operator.")
	flag.StringVar(&operatorNamespace, "operator-namespace", "default", "The namespace of the running operator")
	flag.BoolVar(&createResources, "create-resources", false, "Create resources such as webhooks and configmaps on startup")
	flag.Parse()

	ctrl.SetLogger(zap.Logger(true))

	config := ctrl.GetConfigOrDie()

	//Override operator namespace from environment variable as the source of truth
	operatorNamespace = controllers.GetEnv("POD_NAMESPACE", operatorNamespace)

	watchNamespace := controllers.GetEnv("WATCH_NAMESPACE", "")
	if watchNamespace != "" {
		setupLog.Info("Overriding namespace from WATCH_NAMESPACE", "watchNamespace", watchNamespace)
		namespace = watchNamespace
	}

	if createResources {
		setupLog.Info("Intializing operator")
		err := k8sutils.InitializeOperator(config, operatorNamespace, setupLog, scheme, namespace != "")
		if err != nil {
			setupLog.Error(err, "unable to initialise operator")
			os.Exit(1)
		}
	}

	mgr, err := ctrl.NewManager(config, ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "a33bd623.machinelearning.seldon.io",
		Port:               webHookPort,
		Namespace:          namespace,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.SeldonDeploymentReconciler{
		Client:    mgr.GetClient(),
		Log:       ctrl.Log.WithName("controllers").WithName("SeldonDeployment"),
		Scheme:    mgr.GetScheme(),
		Namespace: namespace,
		Recorder:  mgr.GetEventRecorderFor(constants.ControllerName),
	}).SetupWithManager(mgr, constants.ControllerName); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SeldonDeployment")
		os.Exit(1)
	}

	// Note that we need to create the webhooks for v1alpha2 and v1alpha3 because
	// we are changing our storage version
	if err = (&machinelearningv1alpha2.SeldonDeployment{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "SeldonDeployment")
		os.Exit(1)
	}

	if err = (&machinelearningv1alpha3.SeldonDeployment{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "SeldonDeployment")
		os.Exit(1)
	}

	if err = (&machinelearningv1.SeldonDeployment{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "SeldonDeployment")
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
