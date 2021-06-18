## Upgrading from 1.1 to 1.2 Volume Patch

This notebook contains an overview of how to perform the patch when upgrading from Seldon Core 1.1 into 1.2.

Note that this is ONLY required if you are performing a rolling upgrade. If you can delete the previous version and install Seldon Core 1.2 you will not need to perform any patching.

This issue will be fixed in version 1.2.1, so it is recommended to upgrade to this version instead.

In this notebook we will:
* Install Seldon Core version 1.1
* Deploy 3 models with varying complexities and specifications
* Perform upgrade
* Observe Issues
* Run patch
* Confirm issues are resolved

### Install Seldon Core Version 1.1


```bash
%%bash
kubectl create namespace seldon-system || echo "Namespace seldon-system already exists"
helm upgrade --install seldon-core seldon-core-operator \
    --repo https://storage.googleapis.com/seldon-charts \
    --namespace seldon-system \
    --version v1.1.0 \
    --set certManager.enabled="true" \
    --set usageMetrics.enabled=true \
    --set istio.enabled="true"
```

    Release "seldon-core" does not exist. Installing it now.
    NAME: seldon-core
    LAST DEPLOYED: Sat Jun 27 10:52:41 2020
    NAMESPACE: seldon-system
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None


Check seldon controller manager is running correctly


```python
!kubectl get pods -n seldon-system | grep seldon-controller
```

    seldon-controller-manager-6978f54b99-xvgvd      1/1     Running   0          7m28s


Check no errors in logs


```python
!kubectl logs -n seldon-system -l  control-plane=seldon-controller-manager | tail -2
```

    2020-06-27T09:59:37.767Z	DEBUG	controller-runtime.controller	Successfully Reconciled	{"controller": "seldon-controller-manager", "request": "seldon-system/sklearn"}
    2020-06-27T09:59:37.767Z	DEBUG	controller-runtime.manager.events	Normal	{"object": {"kind":"SeldonDeployment","namespace":"seldon-system","name":"sklearn","uid":"4fca069c-eab1-4903-ad23-40517c91207b","apiVersion":"machinelearning.seldon.io/v1","resourceVersion":"1718083"}, "reason": "Updated", "message": "Updated SeldonDeployment \"sklearn\""}


### Deploy 3 models

#### First model is simple sklearn model in default namespace


```bash
%%bash
kubectl apply -n default -f - << END
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: sklearn
spec:
  name: iris
  predictors:
  - graph:
      children: []
      implementation: SKLEARN_SERVER
      modelUri: gs://seldon-models/sklearn/iris
      name: classifier
    name: default
    replicas: 1
    svcOrchSpec:
      env:
      - name: SELDON_LOG_LEVEL
        value: DEBUG
END
```

    seldondeployment.machinelearning.seldon.io/sklearn created


#### Second model is the same sklaern model but in the seldon-system namespace


```bash
%%bash
kubectl apply -n seldon-system -f - << END
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: sklearn
spec:
  name: iris
  predictors:
  - graph:
      children: []
      implementation: SKLEARN_SERVER
      modelUri: gs://seldon-models/sklearn/iris
      name: classifier
    name: default
    replicas: 1
    svcOrchSpec:
      env:
      - name: SELDON_LOG_LEVEL
        value: DEBUG
END
```

    seldondeployment.machinelearning.seldon.io/sklearn created


#### Third model is the iris custom model with a mounted volume from a secret

First we create the secret


```bash
%%bash 
kubectl create secret generic seldon-test-secret --from-literal=file1.txt=contents --from-literal=file2.txt=morecontents
```

    secret/seldon-test-secret created


Then we deploy the model


```bash
%%bash
kubectl apply -f - << END
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: seldon-deployment-example
spec:
  name: sklearn-iris-deployment
  predictors:
  - componentSpecs:
    - spec:
        volumes:
        - name: "secret-mount"
          volumeSource:
            secret: "seldon-test-secret"
        containers:
        - image: seldonio/sklearn-iris:0.1
          imagePullPolicy: IfNotPresent
          name: sklearn-iris-classifier
          volumeMounts:
          - name: "secret-mount"
            mountPath: "/cert/"
    graph:
      children: []
      endpoint:
        type: REST
      name: sklearn-iris-classifier
      type: MODEL
    name: sklearn-iris-predictor
    replicas: 1
END
```

    seldondeployment.machinelearning.seldon.io/seldon-deployment-example created


Now we wait until they are deployed


```python
!kubectl get sdep --all-namespaces
```

    NAMESPACE       NAME                        AGE
    default         seldon-deployment-example   39s
    default         sklearn                     60s
    seldon-system   sklearn                     55s



```python
!kubectl get pods -n default && kubectl get pods -n seldon-system 
```

    NAME                                                      READY   STATUS    RESTARTS   AGE
    seldon-92a927e5e90d7602e08ba9b9304f70e8-8544bc96d-qkm6x   2/2     Running   0          73s
    sklearn-default-0-classifier-777f84985b-9tj5r             2/2     Running   0          94s
    NAME                                            READY   STATUS    RESTARTS   AGE
    seldon-controller-manager-6978f54b99-xvgvd      1/1     Running   0          6m57s
    sklearn-default-0-classifier-748c59789b-2lnvh   2/2     Running   0          89s


### Perform upgrade to 1.2


```bash
%%bash
helm upgrade --install seldon-core seldon-core-operator \
    --repo https://storage.googleapis.com/seldon-charts \
    --namespace seldon-system \
    --version v1.2.0 \
    --set certManager.enabled="true" \
    --set usageMetrics.enabled=true \
    --set istio.enabled="true"
```

    Namespace seldon-system already exists
    Release "seldon-core" has been upgraded. Happy Helming!
    NAME: seldon-core
    LAST DEPLOYED: Sat Jun 27 11:03:18 2020
    NAMESPACE: seldon-system
    STATUS: deployed
    REVISION: 2
    TEST SUITE: None


    Error from server (AlreadyExists): namespaces "seldon-system" already exists


### Observe error


```python
!kubectl logs -n seldon-system -l  control-plane=seldon-controller-manager | tail -5
```

    k8s.io/apimachinery/pkg/util/wait.JitterUntil
    	/go/pkg/mod/k8s.io/apimachinery@v0.17.2/pkg/util/wait/wait.go:153
    k8s.io/apimachinery/pkg/util/wait.Until
    	/go/pkg/mod/k8s.io/apimachinery@v0.17.2/pkg/util/wait/wait.go:88
    2020-06-27T10:04:01.898Z	DEBUG	controller-runtime.manager.events	Warning	{"object": {"kind":"SeldonDeployment","namespace":"seldon-system","name":"sklearn","uid":"4fca069c-eab1-4903-ad23-40517c91207b","apiVersion":"machinelearning.seldon.io/v1","resourceVersion":"1719032"}, "reason": "InternalError", "message": "Deployment.apps \"sklearn-default-0-classifier\" is invalid: [spec.template.spec.containers[0].volumeMounts[0].name: Not found: \"podinfo\", spec.template.spec.containers[0].volumeMounts[1].mountPath: Invalid value: \"/etc/podinfo\": must be unique]"}


### Run Patch

The error is due a rename on the volumeMounts. We have created the script below which goes through all the seldon deploymetns across all namespaces to rename the volumeMount from podinfo to "seldon-podinfo".

It is recommended to understand this script fully if this is to be run in prodution as it would clash if any existing volume is actually named "podinfo".


```python
%%writefile patch_volumes_1_2.py
#!/usr/bin/env python3

import yaml
import subprocess
import os
import time

def run(cmd: str):
    cmd_arr = cmd.split()
    output = subprocess.Popen(cmd_arr,
       stdout=subprocess.PIPE,
       stderr=subprocess.STDOUT).communicate()
    output_str = [out.decode() for out in output if out]
    return "\n".join(output_str)

def patch_volumes_seldon_1_2():

    namespaces = run("kubectl get ns -o=name")

    for namespace in namespaces.split():
        namespace = namespace.replace("namespace/", "")
        sdeps_raw = run(f"kubectl get sdep -o yaml -n {namespace}")
        sdeps_dict = yaml.safe_load(sdeps_raw)
        sdep_list = sdeps_dict.get("items")
        if sdep_list:
            for sdep in sdep_list:
                name = sdep.get("metadata", {}).get("name")
                print(f"Processing {name} in namespace {namespace}")
                predictors = sdep.get("spec", {}).get("predictors", [])
                for predictor in predictors:
                    for component_spec in predictor.get("componentSpecs", []):
                        for container in component_spec.get("spec", {}).get("containers", []):
                            for volume_mount in container.get("volumeMounts", []):
                                if volume_mount.get("name") == "podinfo":
                                    print("Patching volume")
                                    volume_mount["name"] = "seldon-podinfo"

                with open("seldon_tmp.yaml", "w") as tmp_file:
                    yaml.dump(sdep, tmp_file)
                    run("kubectl apply -f seldon_tmp.yaml")

                print(yaml.dump(sdep))
                os.remove("seldon_tmp.yaml")


if __name__ == "__main__":
    patch_volumes_seldon_1_2()

```

    Overwriting patch_volumes_1_2.py


Run script


```python
!python patch_volumes_1_2.py
```

    Processing seldon-deployment-example in namespace default
    Patching volume
    apiVersion: machinelearning.seldon.io/v1
    kind: SeldonDeployment
    metadata:
      annotations:
        kubectl.kubernetes.io/last-applied-configuration: '{"apiVersion":"machinelearning.seldon.io/v1","kind":"SeldonDeployment","metadata":{"annotations":{},"name":"seldon-deployment-example","namespace":"default"},"spec":{"name":"sklearn-iris-deployment","predictors":[{"componentSpecs":[{"spec":{"containers":[{"image":"seldonio/sklearn-iris:0.1","imagePullPolicy":"IfNotPresent","name":"sklearn-iris-classifier","volumeMounts":[{"mountPath":"/cert/","name":"secret-mount"}]}],"volumes":[{"name":"secret-mount","volumeSource":{"secret":"seldon-test-secret"}}]}}],"graph":{"children":[],"endpoint":{"type":"REST"},"name":"sklearn-iris-classifier","type":"MODEL"},"name":"sklearn-iris-predictor","replicas":1}]}}
    
          '
      creationTimestamp: '2020-06-27T09:58:26Z'
      generation: 1
      name: seldon-deployment-example
      namespace: default
      resourceVersion: '1719036'
      selfLink: /apis/machinelearning.seldon.io/v1/namespaces/default/seldondeployments/seldon-deployment-example
      uid: 8a15eb91-e614-41d9-9d0e-abc191d3a417
    spec:
      name: sklearn-iris-deployment
      predictors:
      - componentSpecs:
        - metadata:
            creationTimestamp: null
          spec:
            containers:
            - image: seldonio/sklearn-iris:0.1
              imagePullPolicy: IfNotPresent
              name: sklearn-iris-classifier
              ports:
              - containerPort: 6000
                name: metrics
                protocol: TCP
              resources: {}
              volumeMounts:
              - mountPath: /cert/
                name: secret-mount
              - mountPath: /etc/podinfo
                name: seldon-podinfo
            volumes:
            - name: secret-mount
        engineResources: {}
        graph:
          endpoint:
            service_host: localhost
            service_port: 9000
            type: REST
          implementation: UNKNOWN_IMPLEMENTATION
          name: sklearn-iris-classifier
          type: MODEL
        labels:
          version: sklearn-iris-predictor
        name: sklearn-iris-predictor
        replicas: 1
        svcOrchSpec: {}
    status:
      address:
        url: http://seldon-deployment-example-sklearn-iris-predictor.default.svc.cluster.local:8000/api/v1.0/predictions
      deploymentStatus:
        seldon-92a927e5e90d7602e08ba9b9304f70e8:
          availableReplicas: 1
          replicas: 1
      description: 'Deployment.apps "seldon-92a927e5e90d7602e08ba9b9304f70e8" is invalid:
        [spec.template.spec.containers[0].volumeMounts[1].name: Not found: "podinfo",
        spec.template.spec.containers[0].volumeMounts[2].mountPath: Invalid value: "/etc/podinfo":
        must be unique]'
      replicas: 1
      serviceStatus:
        seldon-d0934233541ef6b732c88680f8a0e94f:
          httpEndpoint: seldon-d0934233541ef6b732c88680f8a0e94f.default:9000
          svcName: seldon-d0934233541ef6b732c88680f8a0e94f
        seldon-deployment-example-sklearn-iris-predictor:
          grpcEndpoint: seldon-deployment-example-sklearn-iris-predictor.default:5001
          httpEndpoint: seldon-deployment-example-sklearn-iris-predictor.default:8000
          svcName: seldon-deployment-example-sklearn-iris-predictor
      state: Failed
    
    Processing sklearn in namespace default
    Patching volume
    apiVersion: machinelearning.seldon.io/v1
    kind: SeldonDeployment
    metadata:
      annotations:
        kubectl.kubernetes.io/last-applied-configuration: '{"apiVersion":"machinelearning.seldon.io/v1alpha2","kind":"SeldonDeployment","metadata":{"annotations":{},"name":"sklearn","namespace":"default"},"spec":{"name":"iris","predictors":[{"graph":{"children":[],"implementation":"SKLEARN_SERVER","modelUri":"gs://seldon-models/sklearn/iris","name":"classifier"},"name":"default","replicas":1,"svcOrchSpec":{"env":[{"name":"SELDON_LOG_LEVEL","value":"DEBUG"}]}}]}}
    
          '
      creationTimestamp: '2020-06-27T09:58:05Z'
      generation: 1
      name: sklearn
      namespace: default
      resourceVersion: '1719025'
      selfLink: /apis/machinelearning.seldon.io/v1/namespaces/default/seldondeployments/sklearn
      uid: 4f44a5dc-8da4-45ba-8ace-00e51643c7ff
    spec:
      name: iris
      predictors:
      - componentSpecs:
        - metadata:
            creationTimestamp: '2020-06-27T09:58:05Z'
          spec:
            containers:
            - image: seldonio/sklearnserver_rest:0.3
              name: classifier
              ports:
              - containerPort: 6000
                name: metrics
                protocol: TCP
              resources: {}
              volumeMounts:
              - mountPath: /etc/podinfo
                name: seldon-podinfo
        engineResources: {}
        graph:
          endpoint:
            service_host: localhost
            service_port: 9000
            type: REST
          implementation: SKLEARN_SERVER
          modelUri: gs://seldon-models/sklearn/iris
          name: classifier
          type: MODEL
        labels:
          version: default
        name: default
        replicas: 1
        svcOrchSpec:
          env:
          - name: SELDON_LOG_LEVEL
            value: DEBUG
    status:
      address:
        url: http://sklearn-default.default.svc.cluster.local:8000/api/v1.0/predictions
      deploymentStatus:
        sklearn-default-0-classifier:
          availableReplicas: 1
          replicas: 1
      description: 'Deployment.apps "sklearn-default-0-classifier" is invalid: [spec.template.spec.containers[0].volumeMounts[0].name:
        Not found: "podinfo", spec.template.spec.containers[0].volumeMounts[1].mountPath:
        Invalid value: "/etc/podinfo": must be unique]'
      replicas: 1
      serviceStatus:
        sklearn-default:
          grpcEndpoint: sklearn-default.default:5001
          httpEndpoint: sklearn-default.default:8000
          svcName: sklearn-default
        sklearn-default-classifier:
          httpEndpoint: sklearn-default-classifier.default:9000
          svcName: sklearn-default-classifier
      state: Failed
    
    Processing sklearn in namespace seldon-system
    Patching volume
    apiVersion: machinelearning.seldon.io/v1
    kind: SeldonDeployment
    metadata:
      annotations:
        kubectl.kubernetes.io/last-applied-configuration: '{"apiVersion":"machinelearning.seldon.io/v1alpha2","kind":"SeldonDeployment","metadata":{"annotations":{},"name":"sklearn","namespace":"seldon-system"},"spec":{"name":"iris","predictors":[{"graph":{"children":[],"implementation":"SKLEARN_SERVER","modelUri":"gs://seldon-models/sklearn/iris","name":"classifier"},"name":"default","replicas":1,"svcOrchSpec":{"env":[{"name":"SELDON_LOG_LEVEL","value":"DEBUG"}]}}]}}
    
          '
      creationTimestamp: '2020-06-27T09:58:10Z'
      generation: 1
      name: sklearn
      namespace: seldon-system
      resourceVersion: '1719032'
      selfLink: /apis/machinelearning.seldon.io/v1/namespaces/seldon-system/seldondeployments/sklearn
      uid: 4fca069c-eab1-4903-ad23-40517c91207b
    spec:
      name: iris
      predictors:
      - componentSpecs:
        - metadata:
            creationTimestamp: '2020-06-27T09:58:10Z'
          spec:
            containers:
            - image: seldonio/sklearnserver_rest:0.3
              name: classifier
              ports:
              - containerPort: 6000
                name: metrics
                protocol: TCP
              resources: {}
              volumeMounts:
              - mountPath: /etc/podinfo
                name: seldon-podinfo
        engineResources: {}
        graph:
          endpoint:
            service_host: localhost
            service_port: 9000
            type: REST
          implementation: SKLEARN_SERVER
          modelUri: gs://seldon-models/sklearn/iris
          name: classifier
          type: MODEL
        labels:
          version: default
        name: default
        replicas: 1
        svcOrchSpec:
          env:
          - name: SELDON_LOG_LEVEL
            value: DEBUG
    status:
      address:
        url: http://sklearn-default.seldon-system.svc.cluster.local:8000/api/v1.0/predictions
      deploymentStatus:
        sklearn-default-0-classifier:
          availableReplicas: 1
          replicas: 1
      description: 'Deployment.apps "sklearn-default-0-classifier" is invalid: [spec.template.spec.containers[0].volumeMounts[0].name:
        Not found: "podinfo", spec.template.spec.containers[0].volumeMounts[1].mountPath:
        Invalid value: "/etc/podinfo": must be unique]'
      replicas: 1
      serviceStatus:
        sklearn-default:
          grpcEndpoint: sklearn-default.seldon-system:5001
          httpEndpoint: sklearn-default.seldon-system:8000
          svcName: sklearn-default
        sklearn-default-classifier:
          httpEndpoint: sklearn-default-classifier.seldon-system:9000
          svcName: sklearn-default-classifier
      state: Failed
    


### Confirm issues are resolved

We can now check first that all of the containers are running


```python
!kubectl get pods -n default && kubectl get pods -n seldon-system 
```

    NAME                                                       READY   STATUS    RESTARTS   AGE
    seldon-92a927e5e90d7602e08ba9b9304f70e8-6797cc86f7-cv7f9   2/2     Running   0          69s
    sklearn-default-0-classifier-66cf95c445-s6t4x              2/2     Running   0          68s
    NAME                                           READY   STATUS    RESTARTS   AGE
    seldon-controller-manager-7589ff7596-4zqbv     1/1     Running   0          5m2s
    sklearn-default-0-classifier-c86f87c85-xjxf6   2/2     Running   0          68s


And we confirm that there are no longer any errors in the controller manager logs related to the volumeMount


```python

```
