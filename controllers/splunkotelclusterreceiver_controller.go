// Copyright The OpenTelemetry Authors
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

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1beta1 "github.com/signalfx/splunk-otel-operator/api/v1beta1"
)

// SplunkOtelClusterReceiverReconciler reconciles a SplunkOtelClusterReceiver object
type SplunkOtelClusterReceiverReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core.splunk.com,resources=splunkotelclusterreceivers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.splunk.com,resources=splunkotelclusterreceivers/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SplunkOtelClusterReceiver object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.6.4/pkg/reconcile
func (r *SplunkOtelClusterReceiverReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("splunkotelclusterreceiver", req.NamespacedName)

	// your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SplunkOtelClusterReceiverReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1beta1.SplunkOtelClusterReceiver{}).
		Complete(r)
}
