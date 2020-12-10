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
	"context"
	"flag"
	"os"

	"k8s.io/client-go/kubernetes"

	"github.com/seldonio/seldon-core/operator/constants"
	"github.com/seldonio/seldon-core/operator/utils"

	kedav1alpha1 "github.com/kedacore/keda/api/v1alpha1"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	machinelearningv1alpha2 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1alpha2"
	machinelearningv1alpha3 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1alpha3"
	"github.com/seldonio/seldon-core/operator/controllers"
	k8sutils "github.com/seldonio/seldon-core/operator/utils/k8s"
	"go.uber.org/zap"
	istio "istio.io/client-go/pkg/apis/networking/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	zapf "sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

const (
	logLevelEnvVar  = "SELDON_LOG_LEVEL"
	logLevelDefault = "INFO"
	debugEnvVar     = "SELDON_DEBUG"
)

var (
	debugDefault = false

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
	if utils.GetEnv(controllers.ENV_KEDA_ENABLED, "false") == "true" {
		_ = kedav1alpha1.AddToScheme(scheme)
	}
	if utils.GetEnv(controllers.ENV_ISTIO_ENABLED, "false") == "true" {
		_ = istio.AddToScheme(scheme)
	}
	// +kubebuilder:scaffold:scheme
}

func setupLogger(logLevel string, debug bool) {
	level := zap.InfoLevel
	switch logLevel {
	case "DEBUG":
		level = zap.DebugLevel
	case "INFO":
		level = zap.InfoLevel
	case "WARN":
	case "WARNING":
		level = zap.WarnLevel
	case "ERROR":
		level = zap.ErrorLevel
	case "FATAL":
		level = zap.FatalLevel
	}

	atomicLevel := zap.NewAtomicLevelAt(level)

	logger := zapf.New(
		zapf.UseDevMode(debug),
		zapf.Level(&atomicLevel),
	)

	ctrl.SetLogger(logger)
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var webHookPort int
	var namespace string
	var operatorNamespace string
	var createResources bool
	var debug bool
	var logLevel string

	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.IntVar(&webHookPort, "webhook-port", 443, "Webhook server port")
	flag.StringVar(&namespace, "namespace", "", "The namespace to restrict the operator.")
	flag.StringVar(&operatorNamespace, "operator-namespace", "default", "The namespace of the running operator")
	flag.BoolVar(&createResources, "create-resources", false, "Create resources such as webhooks and configmaps on startup")
	flag.BoolVar(
		&debug,
		"debug", utils.GetEnvAsBool(debugEnvVar, debugDefault),
		"Enable debug mode. Logs will be sampled and less structured.",
	)
	flag.StringVar(&logLevel, "log-level", utils.GetEnv(logLevelEnvVar, logLevelDefault), "Log level.")
	flag.Parse()

	ctx := context.Background()

	setupLogger(logLevel, debug)

	config := ctrl.GetConfigOrDie()

	//Override operator namespace from environment variable as the source of truth
	operatorNamespace = utils.GetEnv("POD_NAMESPACE", operatorNamespace)

	watchNamespace := utils.GetEnv("WATCH_NAMESPACE", "")
	if watchNamespace != "" {
		setupLog.Info("Overriding namespace from WATCH_NAMESPACE", "watchNamespace", watchNamespace)
		namespace = watchNamespace
	}

	if createResources {
		setupLog.Info("Intializing operator")
		err := k8sutils.InitializeOperator(ctx, config, operatorNamespace, setupLog, scheme, namespace != "")
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
		ClientSet: kubernetes.NewForConfigOrDie(config),
		Log:       ctrl.Log.WithName("controllers").WithName("SeldonDeployment"),
		Scheme:    mgr.GetScheme(),
		Namespace: namespace,
		Recorder:  mgr.GetEventRecorderFor(constants.ControllerName),
	}).SetupWithManager(ctx, mgr, constants.ControllerName); err != nil {
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
