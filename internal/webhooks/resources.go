package webhooks

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/semconv"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	annotationName = "app.kubernetes.io/name"
	annotationApp  = "app"
)

func serviceName(pod corev1.Pod, resources map[string]string) string {
	if name := pod.Annotations[annotationApp]; name != "" {
		return name
	}
	if name := pod.Annotations[annotationName]; name != "" {
		return name
	}
	if name := resources[string(semconv.K8SDeploymentNameKey)]; name != "" {
		return name
	}
	if name := resources[string(semconv.K8SStatefulSetNameKey)]; name != "" {
		return name
	}
	if name := resources[string(semconv.K8SJobNameKey)]; name != "" {
		return name
	}
	if name := resources[string(semconv.K8SCronJobNameKey)]; name != "" {
		return name
	}
	if name := resources[string(semconv.K8SPodNameKey)]; name != "" {
		return name
	}
	return pod.Spec.Containers[0].Name
}

// createResourceMap creates resource attribute map.
// User defined attributes (in explicitly set env var) have higher precedence.
func (h *handler) createResourceMap(ctx context.Context, ns corev1.Namespace, pod corev1.Pod) (map[string]string, int) {

	k8sResources := map[attribute.Key]string{}
	k8sResources[semconv.K8SNamespaceNameKey] = ns.Name
	k8sResources[semconv.K8SContainerNameKey] = pod.Spec.Containers[0].Name
	// Some fields might be empty - node name, pod name
	// The pod name might be empty if the pod is created form deployment template
	k8sResources[semconv.K8SPodNameKey] = pod.Name
	k8sResources[semconv.K8SPodUIDKey] = string(pod.UID)
	k8sResources[semconv.K8SNodeNameKey] = pod.Spec.NodeName
	h.addParentResourceLabels(ctx, ns, pod.ObjectMeta, k8sResources)

	res := map[string]string{}
	for k, v := range k8sResources {
		if v != "" {
			res[string(k)] = v
		}
	}

	// get existing resources env var and add them to the map
	existingResourceEnvIdx := getIndexOfEnv(pod.Spec.Containers[0].Env, envOTELResourceAttrs)
	if existingResourceEnvIdx > -1 {
		existingResArr := strings.Split(pod.Spec.Containers[0].Env[existingResourceEnvIdx].Value, ",")
		for _, kv := range existingResArr {
			keyValueArr := strings.Split(strings.TrimSpace(kv), "=")
			if len(keyValueArr) != 2 {
				h.logger.Info("found invalid resource attribute", "resource", pod.Name, "attribute", kv)
				continue
			}
			res[keyValueArr[0]] = keyValueArr[1]
		}
	}

	return res, existingResourceEnvIdx
}

func (h *handler) addParentResourceLabels(ctx context.Context, ns corev1.Namespace, objectMeta metav1.ObjectMeta, resources map[attribute.Key]string) {
	for _, owner := range objectMeta.OwnerReferences {
		switch strings.ToLower(owner.Kind) {
		case "replicaset":
			resources[semconv.K8SReplicaSetNameKey] = owner.Name
			resources[semconv.K8SReplicaSetUIDKey] = string(owner.UID)
			// parent of ReplicaSet is e.g. Deployment which we are interested to know
			rs := appsv1.ReplicaSet{}
			// ignore the error. The object might not exist, the error is not important, getting labels is just the best effort
			//nolint:errcheck
			h.client.Get(ctx, types.NamespacedName{
				Namespace: ns.Name,
				Name:      owner.Name,
			}, &rs)
			h.addParentResourceLabels(ctx, ns, rs.ObjectMeta, resources)
		case "deployment":
			resources[semconv.K8SDeploymentNameKey] = owner.Name
			resources[semconv.K8SDeploymentUIDKey] = string(owner.UID)
		case "statefulset":
			resources[semconv.K8SStatefulSetNameKey] = owner.Name
			resources[semconv.K8SStatefulSetUIDKey] = string(owner.UID)
		case "daemonset":
			resources[semconv.K8SDaemonSetNameKey] = owner.Name
			resources[semconv.K8SDaemonSetUIDKey] = string(owner.UID)
		case "job":
			resources[semconv.K8SJobNameKey] = owner.Name
			resources[semconv.K8SJobUIDKey] = string(owner.UID)
		case "cronjob":
			resources[semconv.K8SCronJobNameKey] = owner.Name
			resources[semconv.K8SCronJobUIDKey] = string(owner.UID)
		}
	}
}

func getIndexOfEnv(envs []corev1.EnvVar, name string) int {
	for i := range envs {
		if envs[i].Name == name {
			return i
		}
	}
	return -1
}

func resourceMapToStr(res map[string]string) string {
	kvPairs := make([]string, 0, len(res))
	for k := range res {
		kvPairs = append(kvPairs, fmt.Sprintf("%s=%s", k, res[k]))
	}
	sort.Strings(kvPairs)
	return strings.Join(kvPairs, ",")
}
