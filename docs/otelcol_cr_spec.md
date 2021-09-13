# SplunkOtelAgent Custom Resource Specification

The below `SplunkOtelAgent` custom resource contains all the specification that can be configured. 

```
apiVersion: splunk.com/v1alpha1
kind: SplunkOtelAgent
metadata:
  name: example-splunk-otel-agent 
spec:

  // +optional SplunkRealm is the Splunk APM Realm your Splukn account
  // exists in. For example, us0, us1, etc.
  splunkRealm: <YOUR_SPLUNK_REALM> 

  // +optional ClusterName is the name of the Kubernetes cluster. This
  will be used to identify this cluster in Splunk dashboards.
  clusterName: <YOUR_CLUSTER_NAME>

  agent:
    // +optional Config is the raw JSON to be used as the agent configuration. Refer to the OpenTelemetry Collector documentation for details.
    config:
    
    // +optional Args is the set of arguments to pass to the OpenTelemetry Collector binary
    args:
      metrics-level: detailed
      log-level: debug
    
    // +optional Image indicates the container image to use for the OpenTelemetry Collector.
    image: ""

    // +optional ImagePullPolicy indicates what image pull policy to be used to retrieve the container image to use for the OpenTelemetry Collector.
    imagePullPolicy: ""
    
    // +optional ServiceAccount indicates the name of an existing service account to use with this instance.
    serviceAccount: ""
    
    // +optional VolumeMounts represents the mount points to use in the underlying collector deployment(s)
    volumeMounts: []
    
    // +optional Volumes represents which volumes to use in the underlying collector deployment(s).
    volumes: []
    
    // +optional Ports allows a set of ports to be exposed by the underlying v1.Service. By default, the operator
    // will attempt to infer the required ports by parsing the .Spec.Config property but this property can be
    // used to open aditional ports that can't be inferred by the operator, like for custom receivers.
    ports: []
    
    // +optional ENV vars to set on the OpenTelemetry Collector's Pods. These can then in certain cases be
    // consumed in the config file for the Collector.
    env: []
    
    // +optional Resources to set on the OpenTelemetry Collector pods.
    resources: {}

    // +optional SecurityContext will be set as the container security context.
    securityContext: {}

    // +optional HostNetwork indicates if the pod should run in the host networking namespace.
    hostNetwork: false
    
    // +optional Toleration to schedule OpenTelemetry Collector pods.
    // This is only relevant to daemonsets, statefulsets and deployments
    tolerations: []

  clusterReceiver:
    // +optional Config is the raw JSON to be used as the cluster receiver configuration. Refer to the OpenTelemetry Collector documentation for details.
    config:
    
    // +optional Args is the set of arguments to pass to the OpenTelemetry Collector binary
    args:
      metrics-level: detailed
      log-level: debug
    
    // +optional Replicas is the number of pod instances for the underlying OpenTelemetry Collector
    // Must be 1 for cluster receiver
    replicas: 1
    
    // +optional Image indicates the container image to use for the OpenTelemetry Collector.
    image: ""

    // +optional ImagePullPolicy indicates what image pull policy to be used to retrieve the container image to use for the OpenTelemetry Collector.
    imagePullPolicy: ""
    
    // +optional ServiceAccount indicates the name of an existing service account to use with this instance.
    serviceAccount: ""
    
    // +optional VolumeMounts represents the mount points to use in the underlying collector deployment(s)
    volumeMounts: []
    
    // +optional Volumes represents which volumes to use in the underlying collector deployment(s).
    volumes: []
    
    // +optional Ports allows a set of ports to be exposed by the underlying v1.Service. By default, the operator
    // will attempt to infer the required ports by parsing the .Spec.Config property but this property can be
    // used to open aditional ports that can't be inferred by the operator, like for custom receivers.
    ports: []
    
    // +optional ENV vars to set on the OpenTelemetry Collector's Pods. These can then in certain cases be
    // consumed in the config file for the Collector.
    env: []
    
    // +optional Resources to set on the OpenTelemetry Collector pods.
    resources: {}

    // +optional SecurityContext will be set as the container security context.
    securityContext: {}

    // +optional Toleration to schedule OpenTelemetry Collector pods.
    // This is only relevant to daemonsets, statefulsets and deployments
    tolerations: []

  gateway:
    // +optional Config is the raw JSON to be used as the gateways's configuration. Refer to the OpenTelemetry Collector documentation for details.
    config:
    
    // +optional Args is the set of arguments to pass to the OpenTelemetry Collector binary
    args:
      metrics-level: detailed
      log-level: debug
    
    // +optional Replicas is the number of pod instances for the underlying OpenTelemetry Collector
    replicas: 1
    
    // +optional Image indicates the container image to use for the OpenTelemetry Collector.
    image: ""

    // +optional ImagePullPolicy indicates what image pull policy to be used to retrieve the container image to use for the OpenTelemetry Collector.
    imagePullPolicy: ""
    
    // +optional ServiceAccount indicates the name of an existing service account to use with this instance.
    serviceAccount: ""
    
    // +optional VolumeMounts represents the mount points to use in the underlying collector deployment(s)
    volumeMounts: []
    
    // +optional Volumes represents which volumes to use in the underlying collector deployment(s).
    volumes: []
    
    // +optional Ports allows a set of ports to be exposed by the underlying v1.Service. By default, the operator
    // will attempt to infer the required ports by parsing the .Spec.Config property but this property can be
    // used to open aditional ports that can't be inferred by the operator, like for custom receivers.
    ports: []
    
    // +optional ENV vars to set on the OpenTelemetry Collector's Pods. These can then in certain cases be
    // consumed in the config file for the Collector.
    env: []
    
    // +optional Resources to set on the OpenTelemetry Collector pods.
    resources: {}

    // +optional SecurityContext will be set as the container security context.
    securityContext: {}
    
    // +optional Toleration to schedule OpenTelemetry Collector pods.
    // This is only relevant to daemonsets, statefulsets and deployments
    tolerations: []
```