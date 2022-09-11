# Java Auto Instrumentation with spring-petclinic
In this example we will setup the spring-petclinic project in Kubernetes. We
will then use the operator to auto instrument the java applications from the
spring-petclinic project.

### 1. Complete the [Getting Started](https://github.com/signalfx/splunk-otel-collector-operator/README.md#getting-started) steps

### 2. Deploy the spring-petclinic (cloud version) project
Original Instructions: [spring-petclinic-cloud Setting Things Up In Kubernetes](https://github.com/spring-petclinic/spring-petclinic-cloud#setting-things-up-in-kubernetes)

The steps below are a summary of the original instructions.

#### 2.1 Download the spring-petclinic-cloud repo

```
git clone git@github.com:spring-petclinic/spring-petclinic-cloud.git
```

#### 2.2 Generate a needed wavefront token for spring-petclinic-cloud

```
cd spring-petclinic-cloud/spring-petclinic-api-gateway
mvn spring-boot:run
cd ..
```

The output from the commands above should contain the wavefront token and uri.

```
management.metrics.export.wavefront.api-token=XXXXXXXX-XXXX-XXXX-XXXX-61969fe8f827
management.metrics.export.wavefront.uri=https://wavefront.surf
```

#### 2.3 Setup the spring-petclinic namespace and related resources

```

kubectl apply -f k8s/init-namespace/
kubectl create secret generic wavefront -n spring-petclinic --from-literal=wavefront-url=https://wavefront.surf --from-literal=wavefront-api-token={CHANGEME}
kubectl apply -f k8s/init-services

helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
helm install vets-db-mysql bitnami/mysql --namespace spring-petclinic --set auth.database=service_instance_db
helm install visits-db-mysql bitnami/mysql --namespace spring-petclinic --set auth.database=service_instance_db
helm install customers-db-mysql bitnami/mysql --namespace spring-petclinic --set auth.database=service_instance_db

export REPOSITORY_PREFIX=springcommunity
./scripts/deployToKubernetes.sh
```

#### 2.4 Verify the pet clinic pods are running

```
kubectl get pods -n spring-petclinic
NAME                              READY   STATUS    RESTARTS   AGE
customers-db-mysql-0              1/1     Running   0          3m29s
vets-db-mysql-0                   1/1     Running   0          3m43s
visits-db-mysql-0                 1/1     Running   0          3m37s
wavefront-proxy-959678c94-lmfnx   1/1     Running   0          11s
```

### 3. Instrument the spring-petclinic pods by patching the related deployments
#### 3.1 Add the inject-java annotation to the spring-petclinic pods patching the related deployments

```
kubectl patch deployment api-gateway -p '{"spec": {"template":{"metadata":{"annotations":{"otel.splunk.com/inject-java":"true"}}}} }'
kubectl patch deployment customers-service -p '{"spec": {"template":{"metadata":{"annotations":{"otel.splunk.com/inject-java":"true"}}}} }'
kubectl patch deployment vets-service -p '{"spec": {"template":{"metadata":{"annotations":{"otel.splunk.com/inject-java":"true"}}}} }'
kubectl patch deployment visits-service -p '{"spec": {"template":{"metadata":{"annotations":{"otel.splunk.com/inject-java":"true"}}}} }'
kubectl patch deployment wavefront-proxy -p '{"spec": {"template":{"metadata":{"annotations":{"otel.splunk.com/inject-java":"true"}}}} }'
```

#### 3.2 Verify the spring-petclinic pods are instrumented
If a pod is properly instrumented, it should have a running container that is
using the splunk-otel-instrumentation-java image they should have the pod
annotation "otel.splunk.com/injection-status: success".

 