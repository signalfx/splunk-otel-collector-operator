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

	"github.com/signalfx/splunk-otel-collector-operator/apis/o11y/v1alpha1"
)

// +kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,failurePolicy=ignore,groups="",resources=pods,verbs=create;update,versions=v1,name=mpod.kb.io,sideEffects=none,admissionReviewVersions={v1,v1beta1}
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=list;watch
// +kubebuilder:rbac:groups=o11y.splunk.com,resources=splunkotelagents,verbs=get;list;watch
// +kubebuilder:rbac:groups="apps",resources=replicasets,verbs=get;list;watch

const (
	operatorNamespace           = "splunk-otel-operator-system"
	envSplunkOtelAgent          = "SPLUNK_OTEL_AGENT"
	envOTELServiceName          = "OTEL_SERVICE_NAME"
	envOTELExporterOTLPEndpoint = "OTEL_EXPORTER_OTLP_ENDPOINT"
	envOTELTracesExporter       = "OTEL_TRACES_EXPORTER"
	envOTELResourceAttrs        = "OTEL_RESOURCE_ATTRIBUTES"

	exporterOTLP   = "otlp"
	exporterJaeger = "jaeger-thrift-splunk"

	annotationConfInjectionEnabled = "o11y.splunk.com/inject-config"
	annotationStatus               = "o11y.splunk.com/injection-status"
	annotationReason               = "o11y.splunk.com/injection-reason"
)

type handler struct {
	client  client.Client
	logger  logr.Logger
	decoder *admission.Decoder
}

type config struct {
	exporter string
	endpoint string
}

// NewHandler creates a new WebhookHandler.
func NewHandler(logger logr.Logger, cl client.Client) admission.Handler {
	return &handler{
		client: cl,
		logger: logger,
	}
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

	if !strings.EqualFold(pod.Annotations[annotationConfInjectionEnabled], "true") {
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
		msg := "unable to get splunk agent spec. make sure SpluknOtelAgent is deployed"
		h.logger.Error(err, msg)
		return h.patch(req, pod, errors.New(msg))
	}

	cfg := configFromSpec(spec)

	pod = h.injectConfigIntoPod(ctx, cfg, pod, ns)
	return h.patch(req, pod, nil)
}

func configFromSpec(spec *v1alpha1.SplunkOtelAgentSpec) config {

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

	return cfg
}

func (h *handler) injectConfigIntoPod(ctx context.Context, cfg config, pod corev1.Pod, ns corev1.Namespace) corev1.Pod {
	if len(pod.Spec.Containers) < 1 {
		h.logger.Info("no containers found in pod", "pod", pod.Name)
		return pod
	}

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

	return pod
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

func (h *handler) getAgentSpec(ctx context.Context) (*v1alpha1.SplunkOtelAgentSpec, error) {
	specs := &v1alpha1.SplunkOtelAgentList{}
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
