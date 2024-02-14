/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package main

import (
	"flag"
	"os"

	//+kubebuilder:scaffold:imports
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	mlopscontrollers "github.com/seldonio/seldon-core/operator/v2/controllers/mlops"
	"github.com/seldonio/seldon-core/operator/v2/scheduler"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(mlopsv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var namespace string
	var clusterwide bool
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":4000", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":4001", "The address the probe endpoint binds to.")
	flag.StringVar(&namespace, "namespace", "", "The namespace to restrict the operator.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&clusterwide, "clusterwide", false, "Allow clusterwide operations")
	opts := zap.Options{
		Development: true,
		TimeEncoder: zapcore.ISO8601TimeEncoder,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	logger := zap.New(zap.UseFlagOptions(&opts))
	ctrl.SetLogger(logger)

	watchNamespace := namespace
	if clusterwide {
		watchNamespace = "" // unset namespace so manager watches all namespaces
	}
	setupLog.Info("Starting manager", "clusterwide", clusterwide)
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Namespace:              watchNamespace,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "e98130ae.seldon.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Create scheduler client
	schedulerClient := scheduler.NewSchedulerClient(logger,
		mgr.GetClient(),
		mgr.GetEventRecorderFor("scheduler-client"))

	if err = (&mlopscontrollers.ModelReconciler{
		Client:    mgr.GetClient(),
		Scheme:    mgr.GetScheme(),
		Scheduler: schedulerClient,
		Recorder:  mgr.GetEventRecorderFor("model-controller"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Model")
		os.Exit(1)
	}
	if err = (&mlopscontrollers.ServerReconciler{
		Client:    mgr.GetClient(),
		Scheme:    mgr.GetScheme(),
		Scheduler: schedulerClient,
		Recorder:  mgr.GetEventRecorderFor("server-controller"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Server")
		os.Exit(1)
	}
	if err = (&mlopscontrollers.PipelineReconciler{
		Client:    mgr.GetClient(),
		Scheme:    mgr.GetScheme(),
		Scheduler: schedulerClient,
		Recorder:  mgr.GetEventRecorderFor("pipeline-controller"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Pipeline")
		os.Exit(1)
	}
	if err = (&mlopscontrollers.ServerConfigReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ServerConfig")
		os.Exit(1)
	}
	if err = (&mlopscontrollers.ExperimentReconciler{
		Client:    mgr.GetClient(),
		Scheme:    mgr.GetScheme(),
		Scheduler: schedulerClient,
		Recorder:  mgr.GetEventRecorderFor("pipeline-controller"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Experiment")
		os.Exit(1)
	}
	if err = (&mlopscontrollers.SeldonRuntimeReconciler{
		Client:    mgr.GetClient(),
		Scheme:    mgr.GetScheme(),
		Scheduler: schedulerClient,
		Recorder:  mgr.GetEventRecorderFor("seldonruntime-controller"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SeldonRuntime")
		os.Exit(1)
	}
	if err = (&mlopscontrollers.SeldonConfigReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SeldonConfig")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
