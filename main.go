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

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	authv1alpha1 "github.com/gp42/aws-auth-operator/api/v1alpha1"
	"github.com/gp42/aws-auth-operator/controllers"
	//+kubebuilder:scaffold:imports
)

var (
	scheme                        = runtime.NewScheme()
	setupLog                      = ctrl.Log.WithName("setup")
	flagAwsAuthSyncConfigName     string
	flagSyncInterval              int
	flagAwsAuthConfigMapName      string
	flagAwsAuthConfigMapNamespace string
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(authv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	flag.StringVar(&flagAwsAuthSyncConfigName, "awsauthsyncconfig-name", "default", "Name of AwsAuthSyncConfig to monitor.")
	flag.IntVar(&flagSyncInterval, "sync-interval", 60, "Sync AWS Auth Configuration every X seconds.")
	flag.StringVar(&flagAwsAuthConfigMapName, "aws-auth-cm-name", "aws-auth", "Name of 'aws-auth' configmap.")
	flag.StringVar(&flagAwsAuthConfigMapNamespace, "aws-auth-cm-namespace", "kube-system", "Namespace of 'aws-auth' configmap.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// Using the SDK's default configuration, loading additional config
	// and credentials values from the environment variables, shared
	// credentials, and shared configuration files
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		ctrl.Log.Error(err, "unable to load SDK config: ")
		os.Exit(1)
	}
	iamSvc := iam.NewFromConfig(cfg)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "033562fd.ops42.org",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.AwsAuthSyncConfigReconciler{
		Client:                    mgr.GetClient(),
		Scheme:                    mgr.GetScheme(),
		AwsAuthSyncConfigName:     flagAwsAuthSyncConfigName,
		SyncInterval:              flagSyncInterval,
		AwsAuthConfigMapName:      flagAwsAuthConfigMapName,
		AwsAuthConfigMapNamespace: flagAwsAuthConfigMapNamespace,
		IamClient:                 iamSvc,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AwsAuthSyncConfig")
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
