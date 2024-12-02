package main

import (
	"context"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/abhinav/resource-reaper/pkg/controller"
	"github.com/go-logr/logr"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
}

func main() {
	opts := zap.Options{
		Development: true,
	}
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                              scheme,
		Cache:                               cache.Options{},
		Logger:                              logr.Logger{},
		LeaderElection:                      false,
		LeaderElectionResourceLock:          "",
		LeaderElectionNamespace:             "",
		LeaderElectionID:                    "",
		LeaderElectionConfig:                &rest.Config{},
		LeaderElectionReleaseOnCancel:       false,
		LeaderElectionResourceLockInterface: nil,
		Metrics:                             server.Options{},
		HealthProbeBindAddress:              "",
		ReadinessEndpointName:               "",
		LivenessEndpointName:                "",
		PprofBindAddress:                    "",
		WebhookServer:                       nil,
		BaseContext: func() context.Context {
			return context.Background()
		},
		EventBroadcaster: nil,
		GracefulShutdownTimeout: func() *time.Duration {
			d := time.Duration(0)
			return &d
		}(),
		Controller: config.Controller{},
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controller.ResourceReaper{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("ResourceReaper"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ResourceReaper")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
