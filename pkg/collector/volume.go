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

// Package collector handles the OpenTelemetry Collector.
package collector

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/signalfx/splunk-otel-operator/api/v1alpha1"
	"github.com/signalfx/splunk-otel-operator/internal/config"
	"github.com/signalfx/splunk-otel-operator/pkg/naming"
)

// Volumes builds the volumes for the given instance, including the config map volume.
func Volumes(cfg config.Config, otelcol v1alpha1.SplunkOtelAgent) []corev1.Volume {
	// create one volume per configmap (agent, gateway, clusterreceiver)
	volumes := []corev1.Volume{}

	volumeNames := []string{}
	if !otelcol.Spec.Agent.Disabled {
		volumeNames = append(volumeNames, naming.ConfigMap(otelcol, "agent"))
	}

	if !otelcol.Spec.ClusterReceiver.Disabled {
		volumeNames = append(volumeNames, naming.ConfigMap(otelcol, "cluster-receiver"))
	}

	if !otelcol.Spec.Gateway.Disabled {
		volumeNames = append(volumeNames, naming.ConfigMap(otelcol, "gateway"))
	}

	for _, name := range volumeNames {
		volumes = append(volumes, corev1.Volume{
			Name: naming.ConfigMapVolume(),
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: name},
					Items: []corev1.KeyToPath{{
						Key:  cfg.CollectorConfigMapEntry(),
						Path: cfg.CollectorConfigMapEntry(),
					}},
				},
			},
		})
	}

	// add user specified volumes
	if len(otelcol.Spec.Agent.Volumes) > 0 {
		volumes = append(volumes, otelcol.Spec.Agent.Volumes...)
	}

	return volumes
}
