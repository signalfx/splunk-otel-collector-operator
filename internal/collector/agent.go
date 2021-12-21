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

package collector

import (
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/signalfx/splunk-otel-collector-operator/apis/otel/v1alpha1"
	"github.com/signalfx/splunk-otel-collector-operator/internal/naming"
)

// Agent builds the deployment for the given instance.
func Agent(logger logr.Logger, otelcol v1alpha1.Agent) appsv1.DaemonSet {
	labels := Labels(otelcol)
	labels["app.kubernetes.io/name"] = naming.Agent(otelcol)

	annotations := Annotations(otelcol)

	return appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        naming.Agent(otelcol),
			Namespace:   otelcol.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: otelcol.Annotations,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: ServiceAccountName(otelcol),
					Containers:         []corev1.Container{Container(logger, otelcol.Spec.Agent)},
					Volumes:            Volumes(otelcol.Spec.Agent, naming.ConfigMap(otelcol, "agent")),
					Tolerations:        otelcol.Spec.Agent.Tolerations,
					HostNetwork:        otelcol.Spec.Agent.HostNetwork,
				},
			},
		},
	}
}
