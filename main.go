/*
Copyright 2022 Matt Wise.

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

// Package main is operator-sdk boilerplate
package main

import (
	"flag"
	"os"

	uzap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	crdsv1alpha1 "github.com/diranged/oz/api/v1alpha1"
	"github.com/diranged/oz/controllers"
	zaplogfmt "github.com/jsternberg/zap-logfmt"
	//+kubebuilder:scaffold:imports
)

const (
	defaultReconciliationInterval = 5
	metricsPort                   = 9443
	controllerKey                 = "controller"
	unableToCreateMsg             = "unable to create controller"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(crdsv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var probeAddr string
	var enableLeaderElection bool
	var requestReconciliationInterval int
	// Boilerplate
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	// Custom
	flag.IntVar(&requestReconciliationInterval, "request-reconciliation-interval", defaultReconciliationInterval, "Access Request reconciliation interval (in minutes)")

	// Reconfigure the default logger. Get rid of the JSON log and switch to a LogFmt logger
	configLog := uzap.NewProductionEncoderConfig()

	// Drop the timestamp field - the operator can use `--timestamps` in kubectl to get the timestamp of when the logs
	// were created, we don't need to log them out.
	configLog.TimeKey = zapcore.OmitKey

	// https://sdk.operatorframework.io/docs/building-operators/golang/references/logging/#custom-zap-logger
	logfmtEncoder := zaplogfmt.NewEncoder(configLog)
	opts := zap.Options{
		Development: true,
		Encoder:     logfmtEncoder,
	}

	// Finish the logger setup - mostly boilerplate below
	opts.BindFlags(flag.CommandLine)
	flag.Parse()
	rootLogger := zap.New(zap.UseFlagOptions(&opts))
	ctrl.SetLogger(rootLogger)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   metricsPort,
		HealthProbeBindAddress: probeAddr,

		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		LeaderElection:                enableLeaderElection,
		LeaderElectionID:              "9b20101a.wizardofoz.co",
		LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, unableToCreateMsg)
		os.Exit(1)
	}

	if err = (&controllers.ExecAccessTemplateReconciler{
		OzTemplateReconciler: &controllers.OzTemplateReconciler{
			OzReconciler: &controllers.OzReconciler{
				Client:                  mgr.GetClient(),
				Scheme:                  mgr.GetScheme(),
				APIReader:               mgr.GetAPIReader(),
				ReconcililationInterval: requestReconciliationInterval,
			},
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, unableToCreateMsg, controllerKey, "ExecAccessTemplate")
		os.Exit(1)
	}

	if err = (&controllers.ExecAccessRequestReconciler{
		OzRequestReconciler: &controllers.OzRequestReconciler{
			OzReconciler: &controllers.OzReconciler{
				Client:                  mgr.GetClient(),
				Scheme:                  mgr.GetScheme(),
				APIReader:               mgr.GetAPIReader(),
				ReconcililationInterval: requestReconciliationInterval,
			},
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, unableToCreateMsg, controllerKey, "ExecAccessRequest")
		os.Exit(1)
	}

	if err = (&controllers.AccessTemplateReconciler{
		OzTemplateReconciler: &controllers.OzTemplateReconciler{
			OzReconciler: &controllers.OzReconciler{
				Client:                  mgr.GetClient(),
				Scheme:                  mgr.GetScheme(),
				APIReader:               mgr.GetAPIReader(),
				ReconcililationInterval: requestReconciliationInterval,
			},
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, unableToCreateMsg, controllerKey, "AccessTemplate")
		os.Exit(1)
	}

	if err = (&controllers.AccessRequestReconciler{
		OzRequestReconciler: &controllers.OzRequestReconciler{
			OzReconciler: &controllers.OzReconciler{
				Client:                  mgr.GetClient(),
				Scheme:                  mgr.GetScheme(),
				APIReader:               mgr.GetAPIReader(),
				ReconcililationInterval: requestReconciliationInterval,
			},
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, unableToCreateMsg, controllerKey, "AccessRequest")
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
