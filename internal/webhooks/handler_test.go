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

package webhooks

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/signalfx/splunk-otel-collector-operator/apis/o11y/v1alpha1"
)

func TestConfigFromSpec(t *testing.T) {
	cases := []struct {
		spec *v1alpha1.SplunkOtelAgentSpec
		cfg  config
	}{
		{
			spec: &v1alpha1.SplunkOtelAgentSpec{
				Agent: v1alpha1.SplunkCollectorSpec{},
			},
			cfg: config{
				exporter: "otlp",
				endpoint: "http://$(SPLUNK_OTEL_AGENT):4317",
			},
		},
		{
			spec: &v1alpha1.SplunkOtelAgentSpec{
				Agent: v1alpha1.SplunkCollectorSpec{Disabled: true},
			},
			cfg: config{
				exporter: "otlp",
				endpoint: "http://splunk-otel-collector.splunk-otel-operator-system:4317",
			},
		},
		{
			spec: &v1alpha1.SplunkOtelAgentSpec{
				Agent:   v1alpha1.SplunkCollectorSpec{Disabled: true},
				Gateway: v1alpha1.SplunkCollectorSpec{Disabled: true},
				Realm:   "mars0",
			},
			cfg: config{
				exporter: "jaeger-thrift-splunk",
				endpoint: "https://ingest.mars0.signalfx.com/v2/trace",
			},
		},
	}

	for _, tc := range cases {
		got := configFromSpec(tc.spec)
		assert.Equal(t, tc.cfg, got)
	}
}

func TestInjectConfig(t *testing.T) {
	cases := []struct {
		cfg          config
		container    *corev1.Container
		shouldInject bool
	}{
		{
			cfg:          config{},
			container:    nil,
			shouldInject: false,
		},
		{
			cfg: config{
				exporter: "otlp",
				endpoint: "localhost",
			},
			container: &corev1.Container{
				Name: "test",
			},
			shouldInject: true,
		},
		{
			cfg: config{
				exporter: "some-exporter",
				endpoint: "http://splunk",
			},
			container: &corev1.Container{
				Name: "test",
			},
			shouldInject: true,
		},
	}

	h := &handler{
		logger: logr.DiscardLogger{},
	}

	for _, tc := range cases {
		ns := corev1.Namespace{
			Spec: corev1.NamespaceSpec{},
		}

		pod := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-pod",
			},
			Spec: corev1.PodSpec{Containers: []corev1.Container{}},
		}
		if tc.container != nil {
			pod.Spec.Containers = append(pod.Spec.Containers, *tc.container)
		}
		got := h.injectConfigIntoPod(context.Background(), tc.cfg, pod, ns)

		if !tc.shouldInject {
			continue
		}

		gc := got.Spec.Containers[0]
		assert.Contains(t, gc.Env, corev1.EnvVar{Name: "OTEL_SERVICE_NAME", Value: "test-pod"})
		assert.Contains(t, gc.Env, corev1.EnvVar{Name: "OTEL_TRACES_EXPORTER", Value: tc.cfg.exporter})
		assert.Contains(t, gc.Env, corev1.EnvVar{Name: "OTEL_EXPORTER_OTLP_ENDPOINT", Value: tc.cfg.endpoint})
		assert.Contains(t, gc.Env, corev1.EnvVar{Name: "SPLUNK_OTEL_AGENT", ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "status.hostIP",
			},
		}})
	}
}
