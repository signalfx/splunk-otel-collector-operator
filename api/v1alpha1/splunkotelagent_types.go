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

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SplunkComponentSpec struct {
	// Disabled determines whether this spec will be depoyed or not.
	// +optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Disabled bool `json:"disabled,omitempty"`

	// Config is the raw JSON to be used as the collector's configuration. Refer to the OpenTelemetry Collector documentation for details.
	// +optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Config string `json:"config,omitempty"`

	// Args is the set of arguments to pass to the OpenTelemetry Collector binary
	// +optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Args map[string]string `json:"args,omitempty"`

	// Replicas is the number of pod instances for the underlying OpenTelemetry Collector
	// +optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Replicas *int32 `json:"replicas,omitempty"`

	// ImagePullPolicy indicates the pull policy to be used for retrieving the container image (Always, Never, IfNotPresent)
	// +optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	ImagePullPolicy v1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// TODO(splunk): use correct version number instead of latest
	// Image indicates the container image to use for the OpenTelemetry Collector.
	// +optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Image string `json:"image,omitempty"`

	// ServiceAccount indicates the name of an existing service account to use with this instance.
	// +optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	ServiceAccount string `json:"serviceAccount,omitempty"`

	// SecurityContext will be set as the container security context.
	// +optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	SecurityContext *v1.SecurityContext `json:"securityContext,omitempty"`

	// HostNetwork indicates if the pod should run in the host networking namespace.
	// +optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	HostNetwork bool `json:"hostNetwork,omitempty"`

	// VolumeMounts represents the mount points to use in the underlying collector deployment(s)
	// +optional
	// +listType=atomic
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	VolumeMounts []v1.VolumeMount `json:"volumeMounts,omitempty"`

	// Volumes represents which volumes to use in the underlying collector deployment(s).
	// +optional
	// +listType=atomic
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Volumes []v1.Volume `json:"volumes,omitempty"`

	// Ports allows a set of ports to be exposed by the underlying v1.Service. By default, the operator
	// will attempt to infer the required ports by parsing the .Spec.Config property but this property can be
	// used to open aditional ports that can't be inferred by the operator, like for custom receivers.
	// +optional
	// +listType=atomic
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Ports []v1.ServicePort `json:"ports,omitempty"`

	// ENV vars to set on the OpenTelemetry Collector's Pods. These can then in certain cases be
	// consumed in the config file for the Collector.
	// +optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Env []v1.EnvVar `json:"env,omitempty"`

	// Resources to set on the OpenTelemetry Collector pods.
	// +optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Resources v1.ResourceRequirements `json:"resources,omitempty"`

	// Toleration to schedule OpenTelemetry Collector pods.
	// This is only relevant to daemonsets and deployments
	// +optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
}

// TODO(splunk): rename to SplunkDeploymentSpec or SplunkUnifiedAgentSpec
// SplunkOtelAgentSpec defines the desired state of SplunkOtelAgent.
type SplunkOtelAgentSpec struct {
	// ClusterName is the name of the Kubernetes cluster. This will be used to identify this cluster in Splunk dashboards.
	// +required
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	ClusterName string `json:"clusterName"`

	// SplunkRealm is the Splunk APM Realm your Splukn account exists in. For example, us0, us1, etc.
	// +required
	// +kubebuilder:validation:Required
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	SplunkRealm string `json:"splunkRealm"`

	// TODO(splunk): may be have a common `Env` field here that is injected into all pods(Agent, ClusterReceiver, Gateway)

	// Agent is a Splunk OpenTelemetry Collector instance deployed as an agent on every node.
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Agent SplunkComponentSpec `json:"agent,omitempty"`

	// ClusterReceiver is a single instance Splunk OpenTelemetry Collector deployement used to monitor the entire cluster.
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	ClusterReceiver SplunkComponentSpec `json:"clusterReceiver,omitempty"`

	// ClusterReceiver is a Splunk OpenTelemetry Collector deployement used to export data to Splunk APM.
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Gateway SplunkComponentSpec `json:"gateway,omitempty"`
}

// TODO(splunk): rename to SplunkDeploymentStatus or SplunkUnifiedAgentStatus
// SplunkOtelAgentStatus defines the observed state of SplunkOtelAgent.
type SplunkOtelAgentStatus struct {
	// Replicas is currently not being set and might be removed in the next version.
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// Version of the managed OpenTelemetry Collector (operand)
	// +optional
	Version string `json:"version,omitempty"`

	// Messages about actions performed by the operator on this resource.
	// +optional
	// +listType=atomic
	Messages []string `json:"messages,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=otelcol;otelcols
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.version",description="OpenTelemetry Version"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +operator-sdk:csv:customresourcedefinitions:displayName="OpenTelemetry Collector"

// SplunkOtelAgent is the Schema for the SplunkOtelAgents API.
type SplunkOtelAgent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SplunkOtelAgentSpec   `json:"spec,omitempty"`
	Status SplunkOtelAgentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SplunkOtelAgentList contains a list of SplunkOtelAgent.
type SplunkOtelAgentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SplunkOtelAgent `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SplunkOtelAgent{}, &SplunkOtelAgentList{})
}
