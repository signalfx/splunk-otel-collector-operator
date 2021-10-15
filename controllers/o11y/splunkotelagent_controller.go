// Copyright The OpenTelemetry Authors
// Copyright Splunk Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package o11y

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/signalfx/splunk-otel-collector-operator/apis/o11y/v1alpha1"
	o11yv1alpha1 "github.com/signalfx/splunk-otel-collector-operator/apis/o11y/v1alpha1"
	"github.com/signalfx/splunk-otel-collector-operator/internal/collector/reconcile"
)

// Task represents a reconciliation task to be executed by the reconciler.
type Task struct {
	Name        string
	Do          func(context.Context, reconcile.Params) error
	BailOnError bool
}

// SplunkOtelAgentReconciler reconciles a SplunkOtelAgent object.
type SplunkOtelAgentReconciler struct {
	client.Client
	logger   logr.Logger
	scheme   *runtime.Scheme
	recorder record.EventRecorder
	tasks    []Task
}

// NewReconciler creates a new reconciler for SplunkOtelAgent objects.
func NewReconciler(logger logr.Logger, client client.Client, scheme *runtime.Scheme, recorder record.EventRecorder) *SplunkOtelAgentReconciler {
	tasks := []Task{
		// TODO(splunk): see if we should handle creation of the namespace as well
		// this is tricky as,
		//   - access token secret needs to be in the same namespace
		//   - if we add namespace here then agents will be started before user creates secret
		//   - this means agent will crash until the secret is not created
		// the can lead to confusing behavior so it might be better to have the user
		// create namespace and secret both before creating SplunkOtelAgent
		{
			"config maps",
			reconcile.ConfigMaps,
			true,
		},
		{
			"service accounts",
			reconcile.ServiceAccounts,
			true,
		},
		{
			"services",
			reconcile.Services,
			true,
		},
		{
			"cluster receiver",
			reconcile.ClusterReceivers,
			true,
		},
		{
			"agent",
			reconcile.Agents,
			true,
		},
		/*
			{
				"gateway",
				reconcile.Gateway,
				true,
			},
		*/
		{
			"splunk opentelemetry",
			reconcile.Self,
			true,
		},
	}

	return &SplunkOtelAgentReconciler{
		Client:   client,
		logger:   logger,
		scheme:   scheme,
		recorder: recorder,
		tasks:    tasks,
	}
}

//+kubebuilder:rbac:groups=o11y.splunk.com,resources=splunkotelagents,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=o11y.splunk.com,resources=splunkotelagents/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=o11y.splunk.com,resources=splunkotelagents/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *SplunkOtelAgentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	log := r.logger.WithValues("splunkotelagent", req.NamespacedName)

	var instance v1alpha1.SplunkOtelAgent
	if err := r.Get(ctx, req.NamespacedName, &instance); err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, "unable to fetch SplunkOtelAgent")
		}

		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	params := reconcile.Params{
		Client:   r.Client,
		Instance: instance,
		Log:      log,
		Scheme:   r.scheme,
		Recorder: r.recorder,
	}

	if err := r.RunTasks(ctx, params); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// RunTasks runs all the tasks associated with this reconciler.
func (r *SplunkOtelAgentReconciler) RunTasks(ctx context.Context, params reconcile.Params) error {
	for _, task := range r.tasks {
		if err := task.Do(ctx, params); err != nil {
			r.logger.Error(err, fmt.Sprintf("failed to reconcile %s", task.Name))
			if task.BailOnError {
				return err
			}
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SplunkOtelAgentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&o11yv1alpha1.SplunkOtelAgent{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.Deployment{}).
		Owns(&appsv1.DaemonSet{}).
		Complete(r)
}
