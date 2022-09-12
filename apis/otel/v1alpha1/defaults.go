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

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
)

func newEnvVar(name, value string) v1.EnvVar {
	return v1.EnvVar{
		Name:  name,
		Value: value,
	}
}

func newEnvVarWithFieldRef(name, path string) v1.EnvVar {
	return v1.EnvVar{
		Name: name,
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  path,
			},
		},
	}
}

const (
	defaultAgentCPU    = "200m"
	defaultAgentMemory = "500Mi"
	defaultAgentConfig = `
extensions:
  health_check:
    endpoint: '0.0.0.0:13133'
  k8s_observer:
    auth_type: serviceAccount
    node: '${MY_NODE_NAME}'
  memory_ballast:
    size_mib: ${SPLUNK_BALLAST_SIZE_MIB}
  zpages:
    endpoint: '0.0.0.0:55679'
receivers:
  jaeger:
    protocols:
      grpc:
        endpoint: '0.0.0.0:14250'
      thrift_http:
        endpoint: '0.0.0.0:14268'
  otlp:
    protocols:
      grpc:
        endpoint: '0.0.0.0:4317'
      http:
        endpoint: '0.0.0.0:55681'
  zipkin:
    endpoint: '0.0.0.0:9411'
  smartagent/signalfx-forwarder:
    listenAddress: '0.0.0.0:9080'
    type: signalfx-forwarder
  signalfx:
    endpoint: '0.0.0.0:9943'
  hostmetrics:
    collection_interval: 10s
    scrapers:
      cpu: null
      disk: null
      filesystem: null
      load: null
      memory: null
      network: null
      paging: null
      processes: null
  kubeletstats:
    auth_type: serviceAccount
    collection_interval: 10s
    endpoint: '${MY_NODE_IP}:10250'
    extra_metadata_labels:
      - container.id
    metric_groups:
      - container
      - pod
      - node
  receiver_creator:
    receivers: null
    watch_observers:
      - k8s_observer
  prometheus/self:
    config:
      scrape_configs:
        - job_name: otel-agent
          scrape_interval: 10s
          static_configs:
            - targets:
                - '${MY_POD_IP}:8888'
exporters:
  sapm:
    access_token: '${SPLUNK_ACCESS_TOKEN}'
    endpoint: 'https://ingest.${SPLUNK_REALM}.signalfx.com/v2/trace'
  signalfx:
    access_token: '${SPLUNK_ACCESS_TOKEN}'
    api_url: 'https://api.${SPLUNK_REALM}.signalfx.com'
    ingest_url: 'https://ingest.${SPLUNK_REALM}.signalfx.com'
    sync_host_metadata: true
  splunk_hec:
    token: '${SPLUNK_ACCESS_TOKEN}'
    endpoint: 'https://ingest.${SPLUNK_REALM}.signalfx.com/v1/log'
  logging: null
  logging/debug:
    loglevel: debug
processors:
  k8s_tagger:
    extract:
      metadata:
        - k8s.namespace.name
        - k8s.node.name
        - k8s.pod.name
        - k8s.pod.uid
    filter:
      node: '${MY_NODE_NAME}'
  batch: null
  memory_limiter:
    check_interval: 2s
    limit_mib: '${SPLUNK_MEMORY_LIMIT_MIB}'
  resource:
    attributes:
      - action: insert
        key: k8s.node.name
        value: '${MY_NODE_NAME}'
      - action: insert
        key: k8s.cluster.name
        value: '${MY_CLUSTER_NAME}'
      - action: insert
        key: deployment.environment
        value: '${MY_CLUSTER_NAME}'
  resource/self:
    attributes:
      - action: insert
        key: k8s.pod.name
        value: '${MY_POD_NAME}'
      - action: insert
        key: k8s.pod.uid
        value: '${MY_POD_UID}'
      - action: insert
        key: k8s.namespace.name
        value: '${MY_NAMESPACE}'
  resourcedetection:
    override: false
    timeout: 10s
    detectors:
      - system
      - env
service:
  extensions:
    - health_check
    - k8s_observer
    - memory_ballast
    - zpages
  pipelines:
    traces:
      receivers:
        - smartagent/signalfx-forwarder
        - otlp
        - jaeger
        - zipkin
      processors:
        - k8s_tagger
        - batch
        - resource
        - resourcedetection
      exporters:
        - sapm
        - signalfx
    metrics:
      receivers:
        - hostmetrics
        - kubeletstats
        - receiver_creator
        - signalfx
      processors:
        - batch
        - resource
        - resourcedetection
      exporters:
        - signalfx
    metrics/self:
      receivers:
        - prometheus/self
      processors:
        - batch
        - resource
        - resource/self
        - resourcedetection
      exporters:
        - signalfx
`

	defaultClusterReceiverCPU    = "200m"
	defaultClusterReceiverMemory = "500Mi"
	defaultClusterReceiverConfig = `
extensions:
  health_check:
    endpoint: '0.0.0.0:13133'
  memory_ballast:
    size_mib: ${SPLUNK_BALLAST_SIZE_MIB}
receivers:
  k8s_cluster:
    auth_type: serviceAccount
    metadata_exporters:
      - signalfx
  prometheus/self:
    config:
      scrape_configs:
        - job_name: otel-k8s-cluster-receiver
          scrape_interval: 10s
          static_configs:
            - targets:
                - '${MY_POD_IP}:8888'
exporters:
  signalfx:
    access_token: '${SPLUNK_ACCESS_TOKEN}'
    api_url: 'https://api.${SPLUNK_REALM}.signalfx.com'
    ingest_url: 'https://ingest.${SPLUNK_REALM}.signalfx.com'
    timeout: 10s
  logging: null
  logging/debug:
    loglevel: debug
processors:
  batch: null
  memory_limiter:
    check_interval: 2s
    limit_mib: '${SPLUNK_MEMORY_LIMIT_MIB}'
  resource:
    attributes:
      - action: insert
        key: metric_source
        value: kubernetes
      - action: insert
        key: receiver
        value: k8scluster
      - action: upsert
        key: k8s.cluster.name
        value: '${MY_CLUSTER_NAME}'
      - action: upsert
        key: deployment.environment
        value: '${MY_CLUSTER_NAME}'
  resource/self:
    attributes:
      - action: insert
        key: k8s.node.name
        value: '${MY_NODE_NAME}'
      - action: insert
        key: k8s.pod.name
        value: '${MY_POD_NAME}'
      - action: insert
        key: k8s.pod.uid
        value: '${MY_POD_UID}'
      - action: insert
        key: k8s.namespace.name
        value: '${MY_NAMESPACE}'
  resourcedetection:
    override: false
    timeout: 10s
    detectors:
      - system
      - env
service:
  extensions:
    - health_check
    - memory_ballast
  pipelines:
    metrics:
      receivers:
        - k8s_cluster
      processors:
        - batch
        - resource
        - resourcedetection
      exporters:
        - signalfx
    metrics/self:
      receivers:
        - prometheus/self
      processors:
        - batch
        - resource
        - resource/self
        - resourcedetection
      exporters:
        - signalfx
`

	defaultClusterReceiverConfigOpenshift = `
extensions:
  health_check:
    endpoint: '0.0.0.0:13133'
  memory_ballast:
    size_mib: ${SPLUNK_BALLAST_SIZE_MIB}
receivers:
  k8s_cluster:
    distribution: openshift
    auth_type: serviceAccount
    metadata_exporters:
      - signalfx
  prometheus/self:
    config:
      scrape_configs:
        - job_name: otel-k8s-cluster-receiver
          scrape_interval: 10s
          static_configs:
            - targets:
                - '${MY_POD_IP}:8888'
exporters:
  signalfx:
    access_token: '${SPLUNK_ACCESS_TOKEN}'
    api_url: 'https://api.${SPLUNK_REALM}.signalfx.com'
    ingest_url: 'https://ingest.${SPLUNK_REALM}.signalfx.com'
    timeout: 10s
  logging: null
  logging/debug:
    loglevel: debug
processors:
  batch: null
  memory_limiter:
    check_interval: 2s
    limit_mib: '${SPLUNK_MEMORY_LIMIT_MIB}'
  resource:
    attributes:
      - action: insert
        key: metric_source
        value: kubernetes
      - action: insert
        key: receiver
        value: k8scluster
      - action: upsert
        key: k8s.cluster.name
        value: '${MY_CLUSTER_NAME}'
      - action: upsert
        key: deployment.environment
        value: '${MY_CLUSTER_NAME}'
  resource/self:
    attributes:
      - action: insert
        key: k8s.node.name
        value: '${MY_NODE_NAME}'
      - action: insert
        key: k8s.pod.name
        value: '${MY_POD_NAME}'
      - action: insert
        key: k8s.pod.uid
        value: '${MY_POD_UID}'
      - action: insert
        key: k8s.namespace.name
        value: '${MY_NAMESPACE}'
  resourcedetection:
    override: false
    timeout: 10s
    detectors:
      - system
      - env
service:
  extensions:
    - health_check
    - memory_ballast
  pipelines:
    metrics:
      receivers:
        - k8s_cluster
      processors:
        - batch
        - resource
        - resourcedetection
      exporters:
        - signalfx
    metrics/self:
      receivers:
        - prometheus/self
      processors:
        - batch
        - resource
        - resource/self
        - resourcedetection
      exporters:
        - signalfx
`

	defaultGatewayCPU    = "4"
	defaultGatewayMemory = "8Gi"

	// the javaagent version is managed by the update-javaagent-version.sh script.
	defaultJavaAgentVersion = "v1.14.1"
	defaultJavaAgentImage   = "quay.io/signalfx/splunk-otel-instrumentation-java:" + defaultJavaAgentVersion
)
