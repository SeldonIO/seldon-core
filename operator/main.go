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
	"strings"
	"time"

	//+kubebuilder:scaffold:imports
	zap2 "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

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

const (
	defaultReconcileTimeout = 2 * time.Minute
)

func getLogLevel(l string) zap2.AtomicLevel {
	level, err := zap2.ParseAtomicLevel(strings.ToLower(l))
	if err != nil {
		level = zap2.NewAtomicLevelAt(zapcore.DebugLevel)
	}
	return level
}

func getWatchNamespaceConfig(namespace, watchNamespaces string, clusterwide bool) map[string]cache.Config {
	configs := make(map[string]cache.Config)

	if !clusterwide {
		setupLog.Info("Watching namespace", "namespace", namespace)
		configs[namespace] = cache.Config{}
		return configs
	}

	if watchNamespaces == "" {
		setupLog.Info("Clusterwide mode enabled, watching all namespaces")
		return configs
	}

	setupLog.Info("Clusterwide mode enabled, watching namespaces", "namespaces", watchNamespaces)
	nsSet := make(map[string]struct{})

	for _, ns := range strings.Split(watchNamespaces, ",") {
		ns = strings.TrimSpace(ns)
		if ns != "" {
			nsSet[ns] = struct{}{}
		}
	}

	nsSet[namespace] = struct{}{}
	for ns := range nsSet {
		configs[ns] = cache.Config{}
	}

	return configs
}

func main() {
	var (
		metricsAddr              string
		enableLeaderElection     bool
		probeAddr                string
		namespace                string
		watchNamespaces          string
		clusterwide              bool
		logLevel                 string
		useDeploymentsForServers bool
	)

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":4000", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":4001", "The address the probe endpoint binds to.")
	flag.StringVar(&namespace, "namespace", "", "The namespace to restrict the operator.")
	flag.StringVar(&watchNamespaces, "watch-namespaces", "",
		"Comma separated list of namespaces to watch. "+
			"Only used when --clusterwide is set to true. "+
			"Defaults to all namespaces if not set and --clusterwide is true.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&clusterwide, "clusterwide", false, "Allow clusterwide operations")
	flag.StringVar(&logLevel, "log-level", "debug", "The log level to use for the operator.")
	flag.BoolVar(&useDeploymentsForServers, "use-deployments-for-servers", false, "Use server with deployment instead of statefulset.")

	opts := zap.Options{
		Development: true,
		TimeEncoder: zapcore.ISO8601TimeEncoder,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	opts.Level = getLogLevel(logLevel)
	logger := zap.New(zap.UseFlagOptions(&opts))
	ctrl.SetLogger(logger)

	setupLog.Info("Setting log level", "level", logLevel)

	watchNamespaceConfig := getWatchNamespaceConfig(namespace, watchNamespaces, clusterwide)
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: metricsAddr},
		Cache:                  cache.Options{DefaultNamespaces: watchNamespaceConfig},
		WebhookServer:          webhook.NewServer(webhook.Options{Port: 9443}),
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
		Client:                   mgr.GetClient(),
		Scheme:                   mgr.GetScheme(),
		Scheduler:                schedulerClient,
		Recorder:                 mgr.GetEventRecorderFor("server-controller"),
		UseDeploymentsForServers: useDeploymentsForServers,
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

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
