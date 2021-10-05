  
## **⚠ WARNING: This project is Alpha.**  
### Please do not use in production. Things will break without notice.
  
  
---------
  
# Splunk Operator Connector Operator for Kubernetes

The OpenTelemetry Operator is an implementation of a [Kubernetes Operator](https://coreos.com/operators/).

It helps deploy and manage [Splunk OpenTelemetry Connector](https://github.com/signalfx/splunk-opentelemetry-collector)

## Getting started

### 1. Ensure Cert Manager is installed and available in your cluster
To install the operator in an existing cluster, make sure you have [`cert-manager` installed](https://cert-manager.io/docs/installation/) and run:

### 2. Install the Operator
#### 2.a Kubernetes
```
kubectl apply -f https://github.com/signalfx/splunk-otel-operator/releases/download/v0.0.1/splunk-otel-operator.yaml
```

#### 2.b OpenShift
```
kubectl apply -f https://github.com/signalfx/splunk-otel-operator/releases/download/v0.0.1/splunk-otel-operator-openshift.yaml
```

### 3. Add your Splunk APM token

```
kubectl create secret generic splunk-access-token --namespace splunk-otel-operator-system --from-literal=access-token=SPLUNK_ACCESS_TOKEN
```

### 4. Deploy Splunk OpenTelemetry Connector

Once the `splunk-otel--operator` deployment is ready, create an Splunk OpenTelemetry Collector instance, like:

```console
$ kubectl apply -f - <<EOF
apiVersion: splunk.com/v1alpha1
kind: SplunkOtelAgent
metadata:
  name: splunk-otel
  namespace: splunk-otel-operator-system
spec:
  clusterName: <MY_CLUSTER_NAME>
  splunkRealm: <SPLUNK_REALM>
EOF
```

Replace `MY_CLUSTER_NAME` and `SPLUNK_REALM` with your values.

**_WARNING:_** Until the OpenTelemetry Collector format is stable, changes may be required in the above example to remain
compatible with the latest version of the Splunk OpenTelemetry Operator and Splunk OpenTelemetry Connector.

## Compatibility matrix

### OpenTelemetry Operator vs. Kubernetes

We strive to be compatible with the widest range of Kubernetes versions as possible, but some changes to Kubernetes itself require us to break compatibility with older Kubernetes versions, be it because of code incompatibilities, or in the name of maintainability.

Our promise is that we'll follow what's common practice in the Kubernetes world and support N-2 versions, based on the release date of the OpenTelemetry Operator.

The OpenTelemetry Operator *might* work on versions outside of the given range, but when opening new issues, please make sure to test your scenario on a supported version.

| OpenTelemetry Operator | Kubernetes           |
|------------------------|----------------------|
| v0.1.0                 | v1.20 to v1.22       |

## License
  
[Apache 2.0 License](./LICENSE).

[github-workflow]: https://github.com/signalfx/splunk-otel-operator/actions
[github-workflow-img]: https://github.com/signalfx/splunk-otel-operator/workflows/Continuous%20Integration/badge.svg
[goreport-img]: https://goreportcard.com/badge/github.com/signalfx/splunk-otel-operator
[goreport]: https://goreportcard.com/report/github.com/signalfx/splunk-otel-operator
[godoc-img]: https://godoc.org/github.com/signalfx/splunk-otel-operator?status.svg
[godoc]: https://godoc.org/github.com/signalfx/splunk-otel-operator/pkg/apis/opentelemetry/v1alpha1#SplunkOtelAgent
[code-climate]: https://codeclimate.com/github/signalfx/splunk-otel-operator/maintainability
[code-climate-img]: https://api.codeclimate.com/v1/badges/7bb215eea77fc9c24484/maintainability
[codecov]: https://codecov.io/gh/signalfx/splunk-otel-operator
[codecov-img]: https://codecov.io/gh/signalfx/splunk-otel-operator/branch/main/graph/badge.svg
[contributors]: https://github.com/signalfx/splunk-otel-operator/graphs/contributors
[contributors-img]: https://contributors-img.web.app/image?repo=open-telemetry/opentelemetry-operator
