# Seldon Kafka Integration with KEDA scaling over SSL

In this example we will 

 * run SeldonDeployments for a CIFAR10 Tensorflow model which take their inputs from a Kafka topic and push their outputs to a Kafka topic. 
 * We will scale the Seldon Deployment via KEDA.
 * We will consume/product request over SSL

## Requirements

 * [Install gsutil](https://cloud.google.com/storage/docs/gsutil_install)



```python
!pip install -r requirements.txt
```

## Setup Kafka and KEDA

 * Install Strimzi on cluster via out [playbook](https://github.com/SeldonIO/ansible-k8s-collection/blob/master/playbooks/kafka.yaml)

```
ansible-playbook kafka.yaml 
```

 * [Install KEDA](https://keda.sh/docs/2.6/deploy/) (tested on 2.6.1)
   * See docs for [Kafka Scaler](https://keda.sh/docs/2.6/scalers/apache-kafka/)
   

## Create Kafka Cluster

 * Note tls listener is created with authentication


```python
!cat cluster.yaml
```
!kubectl create -f cluster.yaml -n kafka
## Create Kafka User

This will create a secret called seldon-user in the kafka namespace with cert and key we can use later


```python
!cat user.yaml
```


```python
!kubectl create -f user.yaml -n kafka
```

## Create Topics


```python
res = !kubectl get service seldon-kafka-tls-bootstrap -n kafka -o=jsonpath='{.status.loadBalancer.ingress[0].ip}'
ip = res[0]
%env TLS_BROKER=$ip:9093
```


```python
res = !kubectl get service seldon-kafka-plain-bootstrap -n kafka -o=jsonpath='{.status.loadBalancer.ingress[0].ip}'
ip = res[0]
%env BROKER=$ip:9092
```


```python
%%writefile topics.yaml
apiVersion: kafka.strimzi.io/v1beta2
kind: KafkaTopic
metadata:
  name: cifar10-rest-input
  namespace: kafka
  labels:
    strimzi.io/cluster: "seldon"
spec:
  partitions: 2
  replicas: 1
---
apiVersion: kafka.strimzi.io/v1beta2
kind: KafkaTopic
metadata:
  name: cifar10-rest-output
  namespace: kafka
  labels:
    strimzi.io/cluster: "seldon"
spec:
  partitions: 2
  replicas: 1
```

Create two topics with 2 partitions each. This will allow scaling up to 2 replicas.


```python
!kubectl create -f topics.yaml
```

## Install Seldon

  * [Install Seldon](../install/installation.md)
  * [Follow our docs to intstall the Grafana analytics](https://docs.seldon.ai/seldon-core-1/configuration/integrations/analytics).

## Download Test Request Data
We have two example datasets containing 50,000 requests in tensorflow serving format for CIFAR10. One in JSON format and one as length encoded proto buffers.


```python
!gsutil cp gs://seldon-datasets/cifar10/requests/tensorflow/cifar10_tensorflow.json.gz cifar10_tensorflow.json.gz
!gunzip cifar10_tensorflow.json.gz
```

## Test CIFAR10 REST Model

Upload tensorflow serving rest requests to kafka. This may take some time dependent on your network connection.


```python
!python ../../../util/kafka/test-client.py produce $BROKER cifar10-rest-input --file cifar10_tensorflow.json
```


```python
res = !kubectl get service -n kafka seldon-kafka-tls-bootstrap -o=jsonpath='{.spec.clusterIP}'
ip = res[0]
%env TLS_BROKER_CIP=$ip
```


```python
!kubectl create secret generic keda-enable-tls --from-literal=tls=enable -n kafka
```

## Create Trigger Auth

 * References keda-enable-tls secret
 * References seldon-cluster-ca-cert for ca cert
 * References seldon-user for user certificate


```python
!cat trigger-auth.yaml
```

    apiVersion: keda.sh/v1alpha1
    kind: TriggerAuthentication
    metadata:
      name: seldon-kafka-auth
      namespace: kafka
    spec:
      secretTargetRef:
      - parameter: tls
        name: keda-enable-tls
        key: tls
      - parameter: ca
        name: seldon-cluster-ca-cert
        key: ca.crt
      - parameter: cert
        name: seldon-user
        key: user.crt
      - parameter: key
        name: seldon-user
        key: user.key



```python
!kubectl create -f trigger-auth.yaml -n kafka
```


```python
%%writefile cifar10_rest.yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: tfserving-cifar10
  namespace: kafka
spec:
  protocol: tensorflow
  transport: rest
  serverType: kafka  
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - args: 
          - --port=8500
          - --rest_api_port=8501
          - --model_name=resnet32
          - --model_base_path=gs://seldon-models/tfserving/cifar10/resnet32
          - --enable_batching
          image: tensorflow/serving
          name: resnet32
          ports:
          - containerPort: 8501
            name: http
      kedaSpec:
        pollingInterval: 15
        minReplicaCount: 1
        maxReplicaCount: 2
        triggers:
        - type: kafka
          metadata:
            bootstrapServers: TLS_BROKER_CIP
            consumerGroup: model.tfserving-cifar10.kafka
            lagThreshold: "50"
            topic: cifar10-rest-input
            offsetResetPolicy: latest
            #authMode: sasl_ssl (for latest KEDA - not released yet)
          authenticationRef:
            name: seldon-kafka-auth
    svcOrchSpec:
      env:
      - name: KAFKA_BROKER
        value: TLS_BROKER_CIP
      - name: KAFKA_INPUT_TOPIC
        value: cifar10-rest-input
      - name: KAFKA_OUTPUT_TOPIC
        value: cifar10-rest-output
      - name: KAFKA_SECURITY_PROTOCOL
        value: ssl
      - name: KAFKA_SSL_CA_CERT
        valueFrom:
          secretKeyRef:
            name: seldon-cluster-ca-cert
            key: ca.crt
      - name: KAFKA_SSL_CLIENT_CERT
        valueFrom:
          secretKeyRef:
            name: seldon-user
            key: user.crt
      - name: KAFKA_SSL_CLIENT_KEY
        valueFrom:
          secretKeyRef:
            name: seldon-user
            key: user.key
      - name: KAFKA_SSL_CLIENT_KEY_PASS
        valueFrom:
          secretKeyRef:
            name: seldon-user
            key: user.password
    graph:
      name: resnet32
      type: MODEL
      endpoint:
        service_port: 8501
    name: model
    replicas: 1
```


```python
!cat cifar10_rest.yaml | sed s/TLS_BROKER_CIP/$TLS_BROKER_CIP:9093/ | kubectl apply -f -
```


```python
!kubectl delete -f cifar10_rest.yaml
```


```python

```
