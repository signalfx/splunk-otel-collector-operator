  
## **⚠ WARNING: This project is Alpha.**  
### Please do not use in production. Things will break without notice.
  
  
---------
  
# Splunk OpenTelemetry Collector Operator for Kubernetes

The OpenTelemetry Operator is an implementation of a [Kubernetes Operator](https://coreos.com/operators/).

It helps deploy and manage [Splunk OpenTelemetry Collector](https://github.com/signalfx/splunk-otel-collector)

## Getting started
### 1. Deploy the [cert-manager](https://cert-manager.io/docs/)

```  
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.5.2/cert-manager.yaml
```

The cert-manager adds certificates and certificate issuers as resource types in Kubernetes clusters, and simplifies the process of obtaining, renewing and using those certificates. 


### 2. Deploy the Operator  
#### 2.a Kubernetes

```  
kubectl apply -f https://github.com/signalfx/splunk-otel-collector-operator/releases/latest/download/splunk-otel-operator.yaml  
```  
  
#### 2.b OpenShift

```  
kubectl apply -f https://github.com/signalfx/splunk-otel-collector-operator/releases/latest/download/splunk-otel-operator-openshift.yaml  
```  
  
### 3. Add your Splunk token  
  
```  
kubectl create secret generic splunk-access-token --namespace splunk-otel-operator-system --from-literal=access-token=SPLUNK_ACCESS_TOKEN  
```  
A new users could obtain a token by starting a [Splunk Observability trial](https://www.splunk.com/en_us/download/o11y-cloud-free-trial.html) and following these steps for [creating a token](https://docs.splunk.com/Observability/admin/authentication-tokens/tokens.html).

### 4. Deploy the Splunk OpenTelemetry Collector  
  
Once the `splunk-otel-operator` deployment is ready, create a Splunk OpenTelemetry Collector instance:

  
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

### 4. Verify the cert-manager, operator, and collector are up and running properly.
```
kubectl get pods -n cert-manager
NAME                                       READY   STATUS    RESTARTS   AGE
cert-manager-7c9c58cbcb-jwwkk              1/1     Running   0          5m1s
cert-manager-cainjector-5d88544c9c-chwhr   1/1     Running   0          5m1s
cert-manager-webhook-85f88ffb5b-4hrpb      1/1     Running   0          5m1s
kubectl get pods -n splunk-otel-operator-system
NAME                                                       READY   STATUS    RESTARTS   AGE
splunk-otel-agent-pp8wn                                    1/1     Running   0          68s
splunk-otel-cluster-receiver-8f666b5b8-wbncp               1/1     Running   0          68s
splunk-otel-operator-controller-manager-67b86fcf5c-f2sqq   1/1     Running   0          3m38s
```

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

Then you can automatically instrument it by adding the `otel.splunk.com/inject-java: "true"` annotation to the Pod spec (not the deployment):

```
kubectl patch deployment my-java-app -p '{"spec": {"template":{"metadata":{"annotations":{"otel.splunk.com/inject-java":"true"}}}} }' --namespace my-java-app-namespace
```

It may take a short moment for instrumentation of your the pods to complete. Once a pod has been instrumented, the pod will have the annotation "otel.splunk.com/injection-status: success".

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
