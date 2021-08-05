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
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	uberzap "go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	replicationv1alpha1 "github.com/csi-addons/volume-replication-operator/api/v1alpha1"
	"github.com/csi-addons/volume-replication-operator/controllers"
	"github.com/csi-addons/volume-replication-operator/pkg/config"
	// +kubebuilder:scaffold:imports
)

const (
	// defaultTimeout is default timeout for RPC call
	defaultTimeout = time.Minute
)

var (
	scheme          = runtime.NewScheme()
	setupLog        = ctrl.Log.WithName("setup")
	developmentMode bool
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(replicationv1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var leaderElectionNamespace string
	var enableLeaderElection bool
	var probeAddr string
	var opts zap.Options

	if strings.EqualFold(os.Getenv("DEVELOPMENT_MODE"), "true") {
		developmentMode = true
	}

	cfg := config.NewDriverConfig()

	flag.StringVar(&metricsAddr, "metrics-bind-address", "0", "The address the metric endpoint binds to.")
	flag.StringVar(&leaderElectionNamespace, "leader-election-namespace", "", "Namespace where the leader election resource lives")
	flag.StringVar(&cfg.DriverName, "driver-name", "", "The CSI driver name.")
	flag.StringVar(&cfg.DriverEndpoint, "csi-address", "/run/csi/socket", "Address of the CSI driver socket.")
	flag.DurationVar(&cfg.RPCTimeout, "rpc-timeout", defaultTimeout, "The timeout for RPCs to the CSI driver.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":9998", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	opts = zap.Options{
		ZapOpts: []uberzap.Option{
			uberzap.AddCaller(),
			uberzap.AddStacktrace(uberzap.PanicLevel),
		},
	}

	if developmentMode {
		opts.Development = true
		opts.ZapOpts = []uberzap.Option{
			uberzap.AddCaller(),
			uberzap.AddStacktrace(uberzap.WarnLevel),
		}
	}

	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	err := cfg.Validate()
	if err != nil {
		setupLog.Error(err, "error in driver configuration")
		os.Exit(1)
	}

	if leaderElectionNamespace == "" {
		fmt.Fprintln(os.Stderr, "leader-election-namespace is empty")
		os.Exit(1)
	}
	// unique electionID per operator
	electionID := cfg.DriverName + "volume-replication-" + leaderElectionNamespace

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                     scheme,
		MetricsBindAddress:         metricsAddr,
		Port:                       9443,
		HealthProbeBindAddress:     probeAddr,
		LeaderElectionResourceLock: "leases",
		LeaderElection:             enableLeaderElection,
		LeaderElectionNamespace:    leaderElectionNamespace,
		LeaderElectionID:           electionID,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.VolumeReplicationReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("VolumeReplication"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr, cfg); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "VolumeReplication")
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("health", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("check", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
