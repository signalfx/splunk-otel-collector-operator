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

// Package collector handles the OpenTelemetry Collector.
package collector

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/signalfx/splunk-otel-collector-operator/apis/otel/v1alpha1"
	"github.com/signalfx/splunk-otel-collector-operator/internal/naming"
)

// Volumes builds the volumes for the given instance, including the config map volume.
func Volumes(spec v1alpha1.CollectorSpec, configmap string) []corev1.Volume {
	// create one volume per configmap (agent, gateway, clusterreceiver)
	volumes := []corev1.Volume{}

	volumes = append(volumes, corev1.Volume{
		Name: naming.ConfigMapVolume(),
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				// LocalObjectReference: corev1.LocalObjectReference{Name: naming.ConfigMap(spec)},
				LocalObjectReference: corev1.LocalObjectReference{Name: configmap},
				Items: []corev1.KeyToPath{{
					Key:  "collector.yaml",
					Path: "collector.yaml",
				}},
			},
		},
	})

	// add user specified volumes
	if len(spec.Volumes) > 0 {
		volumes = append(volumes, spec.Volumes...)
	}

	return volumes
}
