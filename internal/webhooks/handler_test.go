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
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/signalfx/splunk-otel-collector-operator/apis/otel/v1alpha1"
)

func TestConfigFromSpec(t *testing.T) {
	cases := []struct {
		spec *v1alpha1.AgentSpec
		cfg  config
	}{
		{
			spec: &v1alpha1.AgentSpec{
				Agent: v1alpha1.CollectorSpec{},
				Instrumentation: v1alpha1.Instrumentation{
					Java: v1alpha1.AutoInstrumentation{
						Image: "quay.io/signalfx/splunk-otel-instrumentation-java:v1.2.3",
					},
				},
			},
			cfg: config{
				exporter:  "otlp",
				endpoint:  "http://$(SPLUNK_OTEL_AGENT):4317",
				javaImage: "quay.io/signalfx/splunk-otel-instrumentation-java:v1.2.3",
			},
		},
		{
			spec: &v1alpha1.AgentSpec{
				Agent: v1alpha1.CollectorSpec{Disabled: true},
				Instrumentation: v1alpha1.Instrumentation{
					Java: v1alpha1.AutoInstrumentation{
						Image: "quay.io/signalfx/splunk-otel-instrumentation-java:v1.6.0",
					},
				},
			},
			cfg: config{
				exporter:  "otlp",
				endpoint:  "http://splunk-otel-collector.splunk-otel-operator-system:4317",
				javaImage: "quay.io/signalfx/splunk-otel-instrumentation-java:v1.6.0",
			},
		},
		{
			spec: &v1alpha1.AgentSpec{
				Agent:   v1alpha1.CollectorSpec{Disabled: true},
				Gateway: v1alpha1.CollectorSpec{Disabled: true},
				Realm:   "mars0",
				Instrumentation: v1alpha1.Instrumentation{
					Java: v1alpha1.AutoInstrumentation{
						Image: "quay.io/signalfx/splunk-otel-instrumentation-java:v2.0",
					},
				},
			},
			cfg: config{
				exporter:  "jaeger-thrift-splunk",
				endpoint:  "https://ingest.mars0.signalfx.com/v2/trace",
				javaImage: "quay.io/signalfx/splunk-otel-instrumentation-java:v2.0",
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
		logger: logr.Discard(),
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
		got, err := h.injectConfig(context.Background(), tc.cfg, pod, ns)

		if !tc.shouldInject {
			continue
		}

		require.NoError(t, err)

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

func TestInjectJava(t *testing.T) {
	cases := []struct {
		cfg          config
		container    *corev1.Container
		shouldInject bool
	}{
		{
			cfg: config{
				exporter:  "otlp",
				endpoint:  "localhost",
				javaImage: "quay.io/signalfx/splunk-otel-instrumentation-java:v2.0",
			},
			container: &corev1.Container{
				Name: "test",
			},
			shouldInject: true,
		},
		{
			cfg: config{
				exporter:  "some-exporter",
				endpoint:  "http://splunk",
				javaImage: "quay.io/signalfx/splunk-otel-instrumentation-java:v1.0",
			},
			container: &corev1.Container{
				Name: "test",
			},
			shouldInject: true,
		},
	}

	h := &handler{
		logger: logr.Discard(),
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
		got, err := h.injectJava(context.Background(), tc.cfg, pod, ns)

		if !tc.shouldInject {
			continue
		}

		require.NoError(t, err)

		require.Len(t, got.Spec.Volumes, 1)
		cv := got.Spec.Volumes[0]
		assert.Equal(t, cv.Name, "splunk-instrumentation")
		assert.Equal(t, cv.VolumeSource.EmptyDir, &corev1.EmptyDirVolumeSource{})

		gc := got.Spec.Containers[0]
		assert.Contains(t, gc.Env, corev1.EnvVar{Name: "OTEL_SERVICE_NAME", Value: "test-pod"})
		assert.Contains(t, gc.Env, corev1.EnvVar{Name: "OTEL_TRACES_EXPORTER", Value: tc.cfg.exporter})
		assert.Contains(t, gc.Env, corev1.EnvVar{Name: "OTEL_EXPORTER_OTLP_ENDPOINT", Value: tc.cfg.endpoint})
		assert.Contains(t, gc.Env, corev1.EnvVar{Name: "SPLUNK_OTEL_AGENT", ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "status.hostIP",
			},
		}})

		require.Len(t, gc.VolumeMounts, 1)
		cvm := gc.VolumeMounts[0]
		assert.Equal(t, cvm.Name, "splunk-instrumentation")

		require.Len(t, got.Spec.InitContainers, 1)
		ic := got.Spec.InitContainers[0]
		assert.Equal(t, ic.Name, "splunk-instrumentation")
		assert.Equal(t, ic.Image, tc.cfg.javaImage)
		assert.Equal(t, ic.Command, []string{"cp", "/splunk-otel-javaagent-all.jar", "/splunk/splunk-otel-javaagent-all.jar"})

		require.Len(t, ic.VolumeMounts, 1)
		v := ic.VolumeMounts[0]
		assert.Equal(t, v.Name, "splunk-instrumentation")
		assert.Equal(t, v.MountPath, "/splunk")
	}
}
