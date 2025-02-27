/*
Copyright 2022.

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

package controller

import (
	"context"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-hclog"
	consulreleasecontrollerv1 "github.com/nicholasjackson/consul-release-controller/kubernetes/controller/api/v1"
	"github.com/nicholasjackson/consul-release-controller/kubernetes/controller/controllers"
	"github.com/nicholasjackson/consul-release-controller/plugins/interfaces"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(consulreleasecontrollerv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

type Kubernetes struct {
	mngr     manager.Manager
	ctx      context.Context
	cancel   context.CancelFunc
	log      hclog.Logger
	provider interfaces.Provider
}

func New(p interfaces.Provider) *Kubernetes {
	ctx, cancelFunc := context.WithCancel(context.Background())

	return &Kubernetes{ctx: ctx, cancel: cancelFunc, log: p.GetLogger().Named("kubernetes-controller"), provider: p}
}

func (k *Kubernetes) Start() error {
	logSink := newSinkLogger(k.log)
	ctrl.SetLogger(logr.New(logSink))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     "0",
		Port:                   19443,
		HealthProbeBindAddress: "0",
		LeaderElection:         false,
		LeaderElectionID:       "4224bb32.nicholasjackson.io",
		WebhookServer:          nil,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.ReleaseReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Provider: k.provider,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Release")
		return err
	}
	//+kubebuilder:scaffold:builder

	setupLog.Info("Starting Kubernetes controller")
	if err := mgr.Start(k.ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		return err
	}

	return nil
}

func (k *Kubernetes) Stop() error {
	k.log.Info("Stopping Kubernetes controller")
	k.cancel()
	return nil
}
