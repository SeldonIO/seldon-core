/*
Copyright 2021.

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

	//+kubebuilder:scaffold:imports
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
	var schedulerHost string
	var schedulerPlaintxtPort int
	var schedulerTLSPort int
	var namespace string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":4000", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":4001", "The address the probe endpoint binds to.")
	flag.StringVar(&namespace, "namespace", "", "The namespace to restrict the operator.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&schedulerHost, "scheduler-host", "0.0.0.0", "Scheduler host")
	flag.IntVar(&schedulerPlaintxtPort, "scheduler-plaintxt-port", 9004, "Scheduler port")
	flag.IntVar(&schedulerTLSPort, "scheduler-tls-port", 9044, "Scheduler port")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	logger := zap.New(zap.UseFlagOptions(&opts))
	ctrl.SetLogger(logger)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Namespace:              namespace,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "e98130ae.seldon.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Create and connect to scheduler
	schedulerClient := scheduler.NewSchedulerClient(logger,
		mgr.GetClient(),
		mgr.GetEventRecorderFor("scheduler-client"))
	err = schedulerClient.ConnectToScheduler(schedulerHost, schedulerPlaintxtPort, schedulerTLSPort)
	if err != nil {
		setupLog.Error(err, "unable to connect to scheduler")
		os.Exit(1)
	}

	// Subscribe the event streams from scheduler
	go func() {
		err := schedulerClient.SubscribeModelEvents(context.Background())
		if err != nil {
			setupLog.Error(err, "Failed to subscribe to scheduler model events")
		}
		os.Exit(1)
	}()
	go func() {
		err := schedulerClient.SubscribeServerEvents(context.Background())
		if err != nil {
			setupLog.Error(err, "Failed to subscribe to scheduler server events")
		}
		os.Exit(1)
	}()
	go func() {
		err := schedulerClient.SubscribePipelineEvents(context.Background())
		if err != nil {
			setupLog.Error(err, "Failed to subscribe to scheduler pipeline events")
		}
		os.Exit(1)
	}()
	go func() {
		err := schedulerClient.SubscribeExperimentEvents(context.Background())
		if err != nil {
			setupLog.Error(err, "Failed to subscribe to scheduler experiment events")
		}
		os.Exit(1)
	}()

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
