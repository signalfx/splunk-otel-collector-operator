  
## **⚠ WARNING: This project is Alpha.**  
### Please do not use in production. Things will break without notice.
  
  
---------
  
# Splunk OpenTelemetry Collector Operator for Kubernetes

The OpenTelemetry Operator is an implementation of a [Kubernetes Operator](https://coreos.com/operators/).

It helps deploy and manage [Splunk OpenTelemetry Collector](https://github.com/signalfx/splunk-otel-collector)

## Getting started

### 1. Ensure Cert Manager is installed and available in your cluster
To install the operator in an existing cluster, make sure you have [`cert-manager` installed](https://cert-manager.io/docs/installation/) and run:

### 2. Install the Operator
#### 2.a Kubernetes
```
kubectl apply -f https://github.com/signalfx/splunk-otel-collector-operator/releases/latest/download/splunk-otel-operator.yaml
```

#### 2.b OpenShift
```
kubectl apply -f https://github.com/signalfx/splunk-otel-collector-operator/releases/latest/download/splunk-otel-operator-openshift.yaml
```

### 3. Add your Splunk APM token

```
kubectl create secret generic splunk-access-token --namespace splunk-otel-operator-system --from-literal=access-token=SPLUNK_ACCESS_TOKEN
```

### 4. Deploy Splunk OpenTelemetry Collector

Once the `splunk-otel--operator` deployment is ready, create an Splunk OpenTelemetry Collector instance, like:

```console
$ kubectl apply -f - <<EOF
apiVersion: otel.splunk.com/v1alpha1
kind: Agent
metadata:
  name: splunk-otel
  namespace: splunk-otel-operator-system
spec:
  clusterName: <MY_CLUSTER_NAME>
  realm: <SPLUNK_REALM>
EOF
```

Replace `MY_CLUSTER_NAME` and `SPLUNK_REALM` with your values.

**_WARNING:_** Until the OpenTelemetry Collector format is stable, changes may be required in the above example to remain
compatible with the latest version of the Splunk OpenTelemetry Operator and Splunk OpenTelemetry Collector.

## Automatically instrumenting k8s pods

This operator can automatically inject configuration and instrumentation agents into Kubernetes pods on demand. In order to do so, you'll need to annotate the pods you want to instrument or auto-configure. For example, if your deployment looks like the following:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-java-app
spec:
  template:
    spec:
      containers:
      - name: my-java-app
        image: my-java-app:latest
```

Then you can automatically instrument it by add `otel.splunk.com/inject-java: "true"` to the Pod spec (not the deployment) so that it would look like the following:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-java-app
spec:
  template:
    metadata:
      annotations:
        otel.splunk.com/inject-java: "true"
    spec:
      containers:
      - name: my-java-app
        image: my-java-app:latest
```

This will automatically inject [Splunk OpenTelemetry Java Agent](github.com/signalfx/splunk-otel-java) into the pod and configure it to send telemetry to the OpenTelemetry agents deployed by the operator.

Right now the following annotations are supported:

### otel.splunk.com/inject-java

When this instrumentation is set to `"true"` on a pod, the operator automatically instruments the pod with the Splunk OpenTelemetry Java agent and configures it to send all telemetry data to the OpenTelemetry agents managed by the operator. 

### otel.splunk.com/inject-config

When this instrumentation is set to `"true"` on a pod, the operator only configures the pod to send all telemetry data to the OpenTelemetry agents managed by the operator. Pods are not instrumented in this case and that is left to the user.


## Compatibility matrix

### OpenTelemetry Operator vs. Kubernetes

We strive to be compatible with the widest range of Kubernetes versions as possible, but some changes to Kubernetes itself require us to break compatibility with older Kubernetes versions, be it because of code incompatibilities, or in the name of maintainability.

Our promise is that we'll follow what's common practice in the Kubernetes world and support N-2 versions, based on the release date of the OpenTelemetry Operator.

The Splunk OpenTelemetry Collector Operator *might* work on versions outside of the given range, but when opening new issues, please make sure to test your scenario on a supported version.

| Operator   | Kubernetes           |
|------------|----------------------|
| v0.0.3     | v1.20 to v1.23       |
| v0.0.4     | v1.23 to v1.25       |

## License
  
[Apache 2.0 License](./LICENSE).

[github-workflow]: https://github.com/signalfx/splunk-otel-collector-operator/actions
[github-workflow-img]: https://github.com/signalfx/splunk-otel-collector-operator/workflows/Continuous%20Integration/badge.svg
[goreport-img]: https://goreportcard.com/badge/github.com/signalfx/splunk-otel-collector-operator
[goreport]: https://goreportcard.com/report/github.com/signalfx/splunk-otel-collector-operator
[godoc-img]: https://godoc.org/github.com/signalfx/splunk-otel-collector-operator?status.svg
[godoc]: https://godoc.org/github.com/signalfx/splunk-otel-collector-operator/pkg/apis/opentelemetry/v1alpha1#SplunkOtelAgent
[code-climate]: https://codeclimate.com/github/signalfx/splunk-otel-operator/maintainability
[code-climate-img]: https://api.codeclimate.com/v1/badges/7bb215eea77fc9c24484/maintainability
[codecov]: https://codecov.io/gh/signalfx/splunk-otel-operator
[codecov-img]: https://codecov.io/gh/signalfx/splunk-otel-operator/branch/main/graph/badge.svg
[contributors]: https://github.com/signalfx/splunk-otel-collector-operator/graphs/contributors
[contributors-img]: https://contributors-img.web.app/image?repo=open-telemetry/opentelemetry-operator

>ℹ️&nbsp;&nbsp;SignalFx was acquired by Splunk in October 2019. See [Splunk SignalFx](https://www.splunk.com/en_us/investor-relations/acquisitions/signalfx.html) for more information.
