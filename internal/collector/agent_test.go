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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/signalfx/splunk-otel-collector-operator/apis/o11y/v1alpha1"
	. "github.com/signalfx/splunk-otel-collector-operator/internal/collector"
)

func TestDaemonSetNewDefault(t *testing.T) {
	// prepare
	otelcol := v1alpha1.SplunkOtelAgent{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-instance",
		},
		Spec: v1alpha1.SplunkOtelAgentSpec{Agent: v1alpha1.SplunkCollectorSpec{
			Tolerations: testTolerationValues,
		}},
	}

	// test
	d := Agent(logger, otelcol)

	// verify
	assert.Equal(t, "my-instance-agent", d.Name)
	assert.Equal(t, "my-instance-agent", d.Labels["app.kubernetes.io/name"])
	assert.Equal(t, "true", d.Annotations["prometheus.io/scrape"])
	assert.Equal(t, "8888", d.Annotations["prometheus.io/port"])
	assert.Equal(t, "/metrics", d.Annotations["prometheus.io/path"])
	assert.Equal(t, testTolerationValues, d.Spec.Template.Spec.Tolerations)

	assert.Len(t, d.Spec.Template.Spec.Containers, 1)

	// none of the default annotations should propagate down to the pod
	assert.Empty(t, d.Spec.Template.Annotations)

	// the pod selector should match the pod spec's labels
	assert.Equal(t, d.Spec.Selector.MatchLabels, d.Spec.Template.Labels)
}

func TestDaemonsetHostNetwork(t *testing.T) {
	// test
	d1 := Agent(logger, v1alpha1.SplunkOtelAgent{
		Spec: v1alpha1.SplunkOtelAgentSpec{},
	})
	assert.False(t, d1.Spec.Template.Spec.HostNetwork)

	// verify custom
	d2 := Agent(logger, v1alpha1.SplunkOtelAgent{
		Spec: v1alpha1.SplunkOtelAgentSpec{Agent: v1alpha1.SplunkCollectorSpec{
			HostNetwork: true,
		}},
	})
	assert.True(t, d2.Spec.Template.Spec.HostNetwork)
}
