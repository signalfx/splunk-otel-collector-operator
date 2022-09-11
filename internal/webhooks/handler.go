package webhooks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/signalfx/splunk-otel-collector-operator/apis/otel/v1alpha1"
)

// +kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,failurePolicy=ignore,groups="",resources=pods,verbs=create;update,versions=v1,name=mpod.kb.io,sideEffects=none,admissionReviewVersions={v1,v1beta1}
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=list;watch
// +kubebuilder:rbac:groups=otel.splunk.com,resources=agents,verbs=get;list;watch
// +kubebuilder:rbac:groups="apps",resources=replicasets,verbs=get;list;watch

const (
	operatorNamespace           = "splunk-otel-operator-system"
	envSplunkOtelAgent          = "SPLUNK_OTEL_AGENT"
	envOTELServiceName          = "OTEL_SERVICE_NAME"
	envOTELExporterOTLPEndpoint = "OTEL_EXPORTER_OTLP_ENDPOINT"
	envOTELTracesExporter       = "OTEL_TRACES_EXPORTER"
	envOTELResourceAttrs        = "OTEL_RESOURCE_ATTRIBUTES"
	envJavaToolsOptions         = "JAVA_TOOL_OPTIONS"

	volumeName        = "splunk-instrumentation"
	initContainerName = "splunk-instrumentation"
	javaJVMArgument   = " -javaagent:/splunk/splunk-otel-javaagent-all.jar"
	exporterOTLP      = "otlp"
	exporterJaeger    = "jaeger-thrift-splunk"

	annotationJava   = "otel.splunk.com/inject-java"
	annotationConfig = "otel.splunk.com/inject-config"
	annotationStatus = "otel.splunk.com/injection-status"
	annotationReason = "otel.splunk.com/injection-reason"
)

type injectFn func(ctx context.Context, cfg config, pod corev1.Pod, ns corev1.Namespace) (corev1.Pod, error)

type handler struct {
	client    client.Client
	logger    logr.Logger
	decoder   *admission.Decoder
	injectMap map[string]injectFn
}

type config struct {
	exporter  string
	endpoint  string
	javaImage string
}

// NewHandler creates a new WebhookHandler.
func NewHandler(logger logr.Logger, cl client.Client) admission.Handler {
	h := &handler{
		client: cl,
		logger: logger,
	}
	h.injectMap = map[string]injectFn{
		annotationJava:   h.injectJava,
		annotationConfig: h.injectConfig,
	}
	return h
}

func (h *handler) patch(req admission.Request, pod corev1.Pod, err error) admission.Response {
	if err != nil {
		pod.Annotations[annotationStatus] = "error"
		pod.Annotations[annotationReason] = err.Error()
	}

	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		h.logger.Error(err, "unable to marshal pod", "pod", pod.Name)
		return admission.Allowed("")
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

func (h *handler) Handle(ctx context.Context, req admission.Request) admission.Response {
	pod := corev1.Pod{}
	err := h.decoder.Decode(req, &pod)
	if err != nil {
		h.logger.Error(err, "unable to decode pod")
		return admission.Errored(http.StatusBadRequest, err)
	}

	if pod.Annotations == nil {
		pod.Annotations = map[string]string{}
	}

	injectFunctions := []injectFn{}
	for ann, fn := range h.injectMap {
		if strings.EqualFold(pod.Annotations[ann], "true") {
			injectFunctions = append(injectFunctions, fn)
		}
	}

	if len(injectFunctions) == 0 {
		return admission.Allowed("")
	}

	if len(pod.Spec.Containers) < 1 {
		h.logger.Info("no containers found in pod", "pod", pod.Name)
		return admission.Allowed("")
	}

	// we use the req.Namespace here because the pod might have not been created yet
	ns := corev1.Namespace{}
	err = h.client.Get(ctx, types.NamespacedName{Name: req.Namespace, Namespace: ""}, &ns)
	if err != nil {
		h.logger.Error(err, "unable to get pod namespace", "namespace", req.Namespace)
		return admission.Errored(http.StatusBadRequest, err)
	}

	spec, err := h.getAgentSpec(ctx)
	if err != nil {
		msg := "unable to get splunk agent spec. make sure SplunkOtelAgent is deployed"
		h.logger.Error(err, msg)
		return h.patch(req, pod, errors.New(msg))
	}

	cfg := configFromSpec(spec)

	for _, fn := range injectFunctions {
		pod, err = fn(ctx, cfg, pod, ns)
		if err != nil {
			return h.patch(req, pod, err)
		}
	}
	pod.Annotations[annotationStatus] = "success"
	return h.patch(req, pod, nil)
}

func configFromSpec(spec *v1alpha1.AgentSpec) config {

	cfg := config{
		exporter: exporterOTLP,
		endpoint: "http://$(SPLUNK_OTEL_AGENT):4317",
	}

	if spec.Agent.Disabled {
		if !spec.Gateway.Disabled {
			cfg.endpoint = fmt.Sprintf("http://splunk-otel-collector.%s:4317", operatorNamespace)
		} else {
			cfg.exporter = exporterJaeger
			cfg.endpoint = fmt.Sprintf("https://ingest.%s.signalfx.com/v2/trace", spec.Realm)
		}
	}

	cfg.javaImage = spec.Instrumentation.Java.Image

	return cfg
}

func (h *handler) injectJava(ctx context.Context, cfg config, pod corev1.Pod, ns corev1.Namespace) (corev1.Pod, error) {
	pod, err := h.injectConfig(ctx, cfg, pod, ns)
	if err != nil {
		return pod, err
	}

	container := &pod.Spec.Containers[0]
	idx := getIndexOfEnv(container.Env, envJavaToolsOptions)
	if idx == -1 {
		container.Env = append(container.Env, corev1.EnvVar{
			Name:  envJavaToolsOptions,
			Value: javaJVMArgument,
		})
	} else {
		if container.Env[idx].ValueFrom != nil {
			msg := fmt.Sprintf("Skipping javaagent injection, the container defines JAVA_TOOL_OPTIONS env var value via ValueFrom for container %s", container.Name)
			h.logger.Info(msg)
			return pod, errors.New(msg)
		}
		container.Env[idx].Value = container.Env[idx].Value + javaJVMArgument
	}
	container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
		Name:      volumeName,
		MountPath: "/splunk",
	})

	pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		}})

	pod.Spec.InitContainers = append(pod.Spec.InitContainers, corev1.Container{
		Name:    initContainerName,
		Image:   cfg.javaImage,
		Command: []string{"cp", "/splunk-otel-javaagent-all.jar", "/splunk/splunk-otel-javaagent-all.jar"},
		VolumeMounts: []corev1.VolumeMount{{
			Name:      volumeName,
			MountPath: "/splunk",
		}},
	})

	return pod, nil
}

func (h *handler) injectConfig(ctx context.Context, cfg config, pod corev1.Pod, ns corev1.Namespace) (corev1.Pod, error) {

	container := &pod.Spec.Containers[0]
	resourceAttrs, resourceEnvIdx := h.createResourceMap(ctx, ns, pod)
	// TODO: some attrs such as node name, pod uid could be empty at this stage
	// so we should use k8s downward API to get read them lazily

	newEnv := []corev1.EnvVar{
		{Name: envSplunkOtelAgent, ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "status.hostIP",
			},
		}},
		{Name: envOTELServiceName, Value: serviceName(pod, resourceAttrs)},
		{Name: envOTELExporterOTLPEndpoint, Value: cfg.endpoint},
		{Name: envOTELTracesExporter, Value: cfg.exporter},
		// TODO: add SPLUNK_ACCESS_TOKEN using env from
		// 1. check if "splunk-access-token" secret key is present in
		// the same namespace as the pod and use ValueFrom to inject it.
		// 2. if it is not present, read it from the operator namespace
		// and inject as an environment variable?
	}

	resourceEnv := corev1.EnvVar{Name: envOTELResourceAttrs, Value: resourceMapToStr(resourceAttrs)}
	if resourceEnvIdx > -1 {
		container.Env[resourceEnvIdx] = resourceEnv
	} else {
		newEnv = append(newEnv, resourceEnv)
	}

	container.Env = h.injectEnvVars(container.Env, newEnv)

	return pod, nil
}

func (h *handler) injectEnvVars(old []corev1.EnvVar, new []corev1.EnvVar) []corev1.EnvVar {
	contains := func(s []corev1.EnvVar, e corev1.EnvVar) bool {
		for _, a := range s {
			if e == a {
				return true
			}
		}
		return false
	}

	for _, ne := range new {
		if !contains(old, ne) {
			old = append(old, ne)
		}
	}
	return old
}

// podAnnotator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (h *handler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

func (h *handler) getAgentSpec(ctx context.Context) (*v1alpha1.AgentSpec, error) {
	specs := &v1alpha1.AgentList{}
	err := h.client.List(ctx, specs)
	if err != nil {
		return nil, err
	}

	switch len(specs.Items) {
	case 0:
		return nil, fmt.Errorf("SplunkOtelAgent is not deployed yet")
	case 1:
		return &specs.Items[0].Spec, nil
	default:
		return nil, fmt.Errorf("found more than one SplunkOtelAgent")
	}
}
