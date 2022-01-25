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
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/signalfx/splunk-otel-collector-operator/internal/autodetect"
)

const (
	defaultJavaAgentImage = "quay.io/signalfx/splunk-otel-instrumentation-java:v1.7.1"
)

// log is for logging in this package.
var agentlog = logf.Log.WithName("agent-resource")

var detectedDistro autodetect.Distro = autodetect.UnknownDistro

func (r *Agent) SetupWebhookWithManager(mgr ctrl.Manager, distro autodetect.Distro) error {
	detectedDistro = distro
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-otel-splunk-com-v1alpha1-agent,mutating=true,failurePolicy=fail,sideEffects=None,groups=otel.splunk.com,resources=agents,verbs=create;update,versions=v1alpha1,name=magent.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &Agent{}

// Default implements webhook.Defaulter so a webhook will be registered for the type.
func (r *Agent) Default() {
	agentlog.Info("default", "name", r.Name)

	if r.Labels == nil {
		r.Labels = map[string]string{}
	}
	if r.Labels["app.kubernetes.io/managed-by"] == "" {
		r.Labels["app.kubernetes.io/managed-by"] = "splunk-otel-collector-operator"
	}

	r.defaultInstrumentation()
	r.defaultAgent()
	r.defaultClusterReceiver()
	r.defaultGateway()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-otel-splunk-com-v1alpha1-agent,mutating=false,failurePolicy=fail,sideEffects=None,groups=otel.splunk.com,resources=agents,verbs=create;update,versions=v1alpha1,name=vagent.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &Agent{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (r *Agent) ValidateCreate() error {
	agentlog.Info("validate create", "name", r.Name)
	return r.validateCRDSpec()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (r *Agent) ValidateUpdate(old runtime.Object) error {
	agentlog.Info("validate update", "name", r.Name)
	return r.validateCRDSpec()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (r *Agent) ValidateDelete() error {
	agentlog.Info("validate delete", "name", r.Name)
	return nil
}

func (r *Agent) validateCRDSpec() error {
	var errs []string

	if err := r.validateInstrumentation(); err != nil {
		errs = append(errs, err.Error())
	}

	if err := r.validateCRDAgentSpec(); err != nil {
		errs = append(errs, err.Error())
	}

	if err := r.validateCRDClusterReceiverSpec(); err != nil {
		errs = append(errs, err.Error())
	}

	if err := r.validateCRDGatewaySpec(); err != nil {
		errs = append(errs, err.Error())
	}
	if len(errs) > 0 {
		return fmt.Errorf(strings.Join(errs, "\n"))
	}
	return nil
}

func (r *Agent) validateInstrumentation() error {
	return nil
}

func (r *Agent) validateCRDAgentSpec() error {
	spec := r.Spec.Agent

	if spec.Replicas != nil {
		return fmt.Errorf("`replicas` is not supported by clusterReceiver")
	}

	return nil
}

func (r *Agent) validateCRDClusterReceiverSpec() error {
	spec := r.Spec.ClusterReceiver

	if spec.Replicas != nil {
		return fmt.Errorf("`replicas` is not supported by clusterReceiver")
	}

	if spec.HostNetwork {
		return fmt.Errorf("`hostNetwork` cannot be true for clusterReceiver")
	}

	return nil
}

func (r *Agent) validateCRDGatewaySpec() error {
	spec := r.Spec.Gateway

	if !r.Spec.Gateway.Disabled {
		return fmt.Errorf("gateway is not supported at the moment")
	}

	if spec.HostNetwork {
		return fmt.Errorf("`hostNetwork` cannot be true for clusterReceiver")
	}

	return nil
}

func (r *Agent) defaultInstrumentation() {
	if r.Spec.Instrumentation.Java.Image == "" {
		r.Spec.Instrumentation.Java.Image = defaultJavaAgentImage
	}
}

func (r *Agent) defaultAgent() {
	realm := r.Spec.Realm
	clusterName := r.Spec.ClusterName

	spec := &r.Spec.Agent
	spec.HostNetwork = true

	if spec.Volumes == nil {
		spec.Volumes = []v1.Volume{
			{
				Name: "hostfs",
				VolumeSource: v1.VolumeSource{
					HostPath: &v1.HostPathVolumeSource{Path: "/"},
				},
			},
			{
				Name: "etc-passwd",
				VolumeSource: v1.VolumeSource{
					HostPath: &v1.HostPathVolumeSource{Path: "/etc/passwd"},
				},
			},
		}
	}

	if spec.VolumeMounts == nil {
		hostToContainer := v1.MountPropagationHostToContainer
		spec.VolumeMounts = []v1.VolumeMount{
			{
				Name:             "hostfs",
				MountPath:        "/hostfs",
				ReadOnly:         true,
				MountPropagation: &hostToContainer,
			},
			{
				Name:      "etc-passwd",
				MountPath: "/etc/passwd",
				ReadOnly:  true,
			},
		}
	}

	if spec.Tolerations == nil {
		spec.Tolerations = []v1.Toleration{
			{
				Key:      "node.alpha.kubernetes.io/role",
				Effect:   v1.TaintEffectNoSchedule,
				Operator: v1.TolerationOpExists,
			},
			{
				Key:      "node-role.kubernetes.io/master",
				Effect:   v1.TaintEffectNoSchedule,
				Operator: v1.TolerationOpExists,
			},
		}
	}

	if spec.Env == nil {
		spec.Env = []v1.EnvVar{
			{
				Name: "SPLUNK_ACCESS_TOKEN",
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{Name: "splunk-access-token"},
						Key:                  "access-token",
					},
				},
			},
			newEnvVar("SPLUNK_REALM", realm),
			newEnvVar("MY_CLUSTER_NAME", clusterName),
			newEnvVar("HOST_PROC", "/hostfs/proc"),
			newEnvVar("HOST_SYS", "/hostfs/sys"),
			newEnvVar("HOST_ETC", "/hostfs/etc"),
			newEnvVar("HOST_VAR", "/hostfs/var"),
			newEnvVar("HOST_RUN", "/hostfs/run"),
			newEnvVar("HOST_DEV", "/hostfs/dev"),
			newEnvVarWithFieldRef("MY_NODE_NAME", "spec.nodeName"),
			newEnvVarWithFieldRef("MY_NODE_IP", "status.hostIP"),
			newEnvVarWithFieldRef("MY_POD_IP", "status.podIP"),
			newEnvVarWithFieldRef("MY_POD_NAME", "metadata.name"),
			newEnvVarWithFieldRef("MY_POD_UID", "metadata.uid"),
			newEnvVarWithFieldRef("MY_NAMESPACE", "metadata.namespace"),
			// TODO(splunk) support ballast
		}
	}

	if spec.Config == "" {
		spec.Config = defaultAgentConfig
	}
}

func (r *Agent) defaultClusterReceiver() {
	realm := r.Spec.Realm
	clusterName := r.Spec.ClusterName

	spec := &r.Spec.ClusterReceiver
	spec.HostNetwork = false

	if spec.Env == nil {
		spec.Env = []v1.EnvVar{
			{
				Name: "SPLUNK_ACCESS_TOKEN",
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{Name: "splunk-access-token"},
						Key:                  "access-token",
					},
				},
			},
			newEnvVar("SPLUNK_REALM", realm),
			newEnvVar("MY_CLUSTER_NAME", clusterName),
			newEnvVar("HOST_PROC", "/hostfs/proc"),
			newEnvVar("HOST_SYS", "/hostfs/sys"),
			newEnvVar("HOST_ETC", "/hostfs/etc"),
			newEnvVar("HOST_VAR", "/hostfs/var"),
			newEnvVar("HOST_RUN", "/hostfs/run"),
			newEnvVar("HOST_DEV", "/hostfs/dev"),
			newEnvVarWithFieldRef("MY_NODE_NAME", "spec.nodeName"),
			newEnvVarWithFieldRef("MY_NODE_IP", "status.hostIP"),
			newEnvVarWithFieldRef("MY_POD_IP", "status.podIP"),
			newEnvVarWithFieldRef("MY_POD_NAME", "metadata.name"),
			newEnvVarWithFieldRef("MY_POD_UID", "metadata.uid"),
			newEnvVarWithFieldRef("MY_NAMESPACE", "metadata.namespace"),
			// TODO(splunk) support ballast
		}
	}

	if spec.Config == "" {
		if detectedDistro == autodetect.OpenShiftDistro {
			spec.Config = defaultClusterReceiverConfigOpenshift
		} else {
			spec.Config = defaultClusterReceiverConfig
		}
	}
}

func (r *Agent) defaultGateway() {
	spec := &r.Spec.Gateway
	// TODO(splunk): forcibly disable gateway until we add support for it.
	spec.Disabled = true
	spec.HostNetwork = false
}
