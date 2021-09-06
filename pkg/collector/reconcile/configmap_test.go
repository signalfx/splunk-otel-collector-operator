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

package reconcile

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"

	"github.com/signalfx/splunk-otel-operator/api/v1alpha1"
	"github.com/signalfx/splunk-otel-operator/internal/config"
)

func TestDefaultConfigMap(t *testing.T) {
	expectedLables := map[string]string{
		"app.kubernetes.io/managed-by": "splunk-otel-operator",
		"app.kubernetes.io/instance":   "default.test",
		"app.kubernetes.io/part-of":    "opentelemetry",
	}

	t.Run("should return expected collector config map", func(t *testing.T) {
		expectedLables["app.kubernetes.io/component"] = "splunk-otel-collector"
		expectedLables["app.kubernetes.io/name"] = "test-collector"

		expectedData := map[string]string{
			"collector.yaml": `processors:
receivers:
  jaeger:
    protocols:
      grpc:
  prometheus:
    config:
      scrape_configs:
        job_name: otel-collector
        scrape_interval: 10s
        static_configs:
          - targets: [ '0.0.0.0:8888', '0.0.0.0:9999' ]

exporters:
  logging:

service:
  pipelines:
    metrics:
      receivers: [prometheus]
      processors: []
      exporters: [logging]`,
		}

		actual := desiredConfigMap(context.Background(), params())

		assert.Equal(t, "test-collector", actual.Name)
		assert.Equal(t, expectedLables, actual.Labels)
		assert.Equal(t, expectedData, actual.Data)

	})
}

func TestDesiredConfigMap(t *testing.T) {
	expectedLables := map[string]string{
		"app.kubernetes.io/managed-by": "splunk-otel-operator",
		"app.kubernetes.io/instance":   "default.test",
		"app.kubernetes.io/part-of":    "opentelemetry",
	}

	t.Run("should return expected collector config map", func(t *testing.T) {
		expectedLables["app.kubernetes.io/component"] = "splunk-otel-collector"
		expectedLables["app.kubernetes.io/name"] = "test-collector"

		expectedData := map[string]string{
			"collector.yaml": `processors:
receivers:
  jaeger:
    protocols:
      grpc:
  prometheus:
    config:
      scrape_configs:
        job_name: otel-collector
        scrape_interval: 10s
        static_configs:
          - targets: [ '0.0.0.0:8888', '0.0.0.0:9999' ]

exporters:
  logging:

service:
  pipelines:
    metrics:
      receivers: [prometheus]
      processors: []
      exporters: [logging]`,
		}

		actual := desiredConfigMap(context.Background(), params())

		assert.Equal(t, "test-collector", actual.Name)
		assert.Equal(t, expectedLables, actual.Labels)
		assert.Equal(t, expectedData, actual.Data)

	})
}

func TestExpectedConfigMap(t *testing.T) {
	t.Run("should create collector config maps", func(t *testing.T) {
		configMap, err := desiredTAConfigMap(params())
		assert.NoError(t, err)
		err = expectedConfigMaps(context.Background(), params(), []v1.ConfigMap{desiredConfigMap(context.Background(), params()), configMap}, true)
		assert.NoError(t, err)

		exists, err := populateObjectIfExists(t, &v1.ConfigMap{}, types.NamespacedName{Namespace: "default", Name: "test-collector"})

		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("should update collector config map", func(t *testing.T) {

		param := Params{
			Config: config.New(),
			Client: k8sClient,
			Instance: v1alpha1.SplunkOtelAgent{
				TypeMeta: metav1.TypeMeta{
					Kind:       "splunk.com",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
					UID:       instanceUID,
				},
			},
			Scheme:   testScheme,
			Log:      logger,
			Recorder: record.NewFakeRecorder(10),
		}
		cm := desiredConfigMap(context.Background(), param)
		createObjectIfNotExists(t, "test-collector", &cm)

		err := expectedConfigMaps(context.Background(), params(), []v1.ConfigMap{desiredConfigMap(context.Background(), params())}, true)
		assert.NoError(t, err)

		actual := v1.ConfigMap{}
		exists, err := populateObjectIfExists(t, &actual, types.NamespacedName{Namespace: "default", Name: "test-collector"})

		assert.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, instanceUID, actual.OwnerReferences[0].UID)
		assert.Equal(t, params().Instance.Spec.Config, actual.Data["collector.yaml"])
	})

	t.Run("should delete config map", func(t *testing.T) {

		deletecm := v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-delete-collector",
				Namespace: "default",
				Labels: map[string]string{
					"app.kubernetes.io/instance":   "default.test",
					"app.kubernetes.io/managed-by": "splunk-otel-operator",
				},
			},
		}
		createObjectIfNotExists(t, "test-delete-collector", &deletecm)

		exists, _ := populateObjectIfExists(t, &v1.ConfigMap{}, types.NamespacedName{Namespace: "default", Name: "test-delete-collector"})
		assert.True(t, exists)

		err := deleteConfigMaps(context.Background(), params(), []v1.ConfigMap{desiredConfigMap(context.Background(), params())})
		assert.NoError(t, err)

		exists, _ = populateObjectIfExists(t, &v1.ConfigMap{}, types.NamespacedName{Namespace: "default", Name: "test-delete-collector"})
		assert.False(t, exists)
	})
}
