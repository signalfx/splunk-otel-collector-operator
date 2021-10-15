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

package collector_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"

	"github.com/signalfx/splunk-otel-collector-operator/apis/o11y/v1alpha1"
	. "github.com/signalfx/splunk-otel-collector-operator/internal/collector"
	"github.com/signalfx/splunk-otel-collector-operator/internal/naming"
)

func TestVolumeNewDefault(t *testing.T) {
	// prepare
	otelcol := v1alpha1.SplunkOtelAgent{}

	// test
	volumes := Volumes(otelcol.Spec.Agent, "splunk-agent")

	// verify
	assert.Len(t, volumes, 1)

	// check that it's the otc-internal volume, with the config map
	assert.Equal(t, naming.ConfigMapVolume(), volumes[0].Name)
}

func TestVolumeAllowsMoreToBeAdded(t *testing.T) {
	// prepare
	otelcol := v1alpha1.SplunkOtelAgent{
		Spec: v1alpha1.SplunkOtelAgentSpec{
			Agent: v1alpha1.SplunkCollectorSpec{
				Volumes: []corev1.Volume{{
					Name: "my-volume",
				}},
			},
		},
	}

	// test
	volumes := Volumes(otelcol.Spec.Agent, "splunk-agent")

	// verify
	assert.Len(t, volumes, 2)

	// check that it's the otc-internal volume, with the config map
	assert.Equal(t, "my-volume", volumes[1].Name)
}
