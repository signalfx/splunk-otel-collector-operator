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

// Gateway builds the Splunk Otel Collector Gateway deployment for the given instance.
func Gateway(logger logr.Logger, otelcol v1alpha1.Agent) appsv1.Deployment {
	labels := Labels(otelcol)
	labels["app.kubernetes.io/name"] = naming.Gateway(otelcol)

	annotations := Annotations(otelcol)

	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        naming.Gateway(otelcol),
			Namespace:   otelcol.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: otelcol.Spec.Gateway.Replicas,
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
					Containers:         []corev1.Container{Container(logger, otelcol.Spec.Gateway)},
					Volumes:            Volumes(otelcol.Spec.Gateway, naming.ConfigMap(otelcol, "gateway")),
					Tolerations:        otelcol.Spec.Gateway.Tolerations,
				},
			},
		},
	}
}
