# Seldon Kafka Integration Example with SpaCy Reddit Model

In this model we will build upon the code from the Seldon SpaCy NLP example for text classification.

You will learn how to deploy the model using the Kafka protocol.

## Requirements

In your commands:
* helm 3.x
* kubectl

In cluster:
* [Install Seldon](../install/installation.md)
* Install Kafka as per the instructions below


## Setup Kafka

Ensure your helm repo has access to the strimzi Kafka charts which we'll use to install.


```python
!helm repo add strimzi https://strimzi.io/charts/
```

    "strimzi" has been added to your repositories


Install the Kafka operator in your cluster


```python
!helm install my-release strimzi/strimzi-kafka-operator
```

    NAME: my-release
    LAST DEPLOYED: Tue Sep 29 08:31:41 2020
    NAMESPACE: default
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None
    NOTES:
    Thank you for installing strimzi-kafka-operator-0.19.0
    
    To create a Kafka cluster refer to the following documentation.
    
    https://strimzi.io/docs/operators/0.19.0/using.html#deploying-cluster-operator-helm-chart-str


We now create a kafka cluster instantiation with a simple setup as outlined below.


```python
%%writefile kafka-cluster-config.yaml
apiVersion: kafka.strimzi.io/v1beta1
kind: Kafka
metadata:
  name: my-cluster
spec:
  kafka:
    replicas: 1
    listeners:
      plain: {}
      tls: {}
      external:
        type: nodeport
        tls: false
    storage:
      type: ephemeral
    config:
      offsets.topic.replication.factor: 1
      transaction.state.log.replication.factor: 1
      transaction.state.log.min.isr: 1
  zookeeper:
    replicas: 1
    storage:
      type: ephemeral
  entityOperator:
    topicOperator: {}
    userOperator: {}
```

    Overwriting kafka-cluster-config.yaml



```python
!kubectl apply -f kafka-cluster-config.yaml
```

    kafka.kafka.strimzi.io/my-cluster created


We can now check that kafka was installed correctly


```python
!kubectl get pods | grep my-cluster
```

    my-cluster-entity-operator-df58f8b9f-s9wx5               3/3     Running   0          148m
    my-cluster-kafka-0                                       2/2     Running   0          149m
    my-cluster-zookeeper-0                                   1/1     Running   0          149m


### Create topics 
We now need to create the input and output topics for our reddit classifier


```python
%%writefile topics.yaml
apiVersion: kafka.strimzi.io/v1beta1
kind: KafkaTopic
metadata:
  name: reddit-classifier-input
  labels:
    strimzi.io/cluster: "my-cluster"
spec:
  partitions: 2
  replicas: 1
---
apiVersion: kafka.strimzi.io/v1beta1
kind: KafkaTopic
metadata:
  name: reddit-classifier-output
  labels:
    strimzi.io/cluster: "my-cluster"
spec:
  partitions: 2
  replicas: 1
```

    Writing topics.yaml



```python
!kubectl apply -f topics.yaml
```

    kafkatopic.kafka.strimzi.io/reddit-classifier-input created
    kafkatopic.kafka.strimzi.io/reddit-classifier-output created


## Train Spacy Sklearn Model

To train the spacy sklearn model you can follow the instructions in the SKlearn Spacy Model Example.

Alternatively you can just use the image that is saved in the seldon dockerhub with the image `seldonio/reddit-classifier:0.1`

## Deploy SpaCy Text Classifier

Now we're able to define the YAML that will be used to deploy the configuration.


```python
%%writefile sdep_reddit_kafka.yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: reddit-kafka
spec:
  serverType: kafka
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/reddit-classifier:0.1
          name: classifier
    svcOrchSpec:
      env:
      - name: KAFKA_BROKER
        value: my-cluster-kafka-brokers.default.svc.cluster.local:9092
      - name: KAFKA_INPUT_TOPIC
        value: reddit-classifier-input
      - name: KAFKA_OUTPUT_TOPIC
        value: reddit-classifier-output
    graph:
      name: classifier
      type: MODEL
    name: default
    replicas: 1
```

    Overwriting sdep_reddit_kafka.yaml



```python
!kubectl apply -f sdep_reddit_kafka.yaml
```

    seldondeployment.machinelearning.seldon.io/reddit-kafka configured


We can confirm that now the model is running as expected:


```python
!kubectl get pods | grep reddit
```

    reddit-kafka-default-0-classifier-c6ccdd66f-vmf4v        2/2     Running   0          45m


## Send real time data for stream processing

We can now send real time data for stream processing. 

Below we send a single input with the text "This is an input", which will be consumed from the input topic.


```python
!kubectl run --quiet=true -it --rm kafkaconsumer --image=bitnami/kafka:2.6.0 --restart=Never --command -- \
    bash -c "echo '{\"data\": {\"ndarray\": [\"This is an input\"]}}' \
    | kafka-console-producer.sh --broker-list my-cluster-kafka-external-bootstrap.default:9094 --topic reddit-classifier-input"
```

### Check the data processed

We now are able to see all the data that has been pushed to the output topic `reddit-classifier-output`.

This allows us to see all the inputs that have been processed. We will be listening for 10 seconds to ensure all data is found.


```python
!kubectl run --quiet=true -it --rm kafkaproducer --image=bitnami/kafka:2.6.0 --restart=Never --command -- \
    kafka-console-consumer.sh --bootstrap-server my-cluster-kafka-external-bootstrap.default:9094 --topic reddit-classifier-output \
        --from-beginning --timeout-ms 10000
```

    {"data":{"names":["t:0","t:1"],"ndarray":[[0.6758450844706712,0.32415491552932885]]},"meta":{}}
    
    {"data":{"names":["t:0","t:1"],"ndarray":[[0.6758450844706712,0.32415491552932885]]},"meta":{}}
    
    {"data":{"names":["t:0","t:1"],"ndarray":[[0.6758450844706712,0.32415491552932885]]},"meta":{}}
    
    {"data":{"names":["t:0","t:1"],"ndarray":[[0.6758450844706712,0.32415491552932885]]},"meta":{}}
    
    {"data":{"names":["t:0","t:1"],"ndarray":[[0.6758450844706712,0.32415491552932885]]},"meta":{}}
    
    {"data":{"names":["t:0","t:1"],"ndarray":[[0.6758450844706712,0.32415491552932885]]},"meta":{}}
    
    {"data":{"names":["t:0","t:1"],"ndarray":[[0.6758450844706712,0.32415491552932885]]},"meta":{}}
    
    {"data":{"names":["t:0","t:1"],"ndarray":[[0.6758450844706712,0.32415491552932885]]},"meta":{}}
    
    {"data":{"names":["t:0","t:1"],"ndarray":[[0.6758450844706712,0.32415491552932885]]},"meta":{}}
    
    [2020-09-29 11:45:06,366] ERROR Error processing message, terminating consumer process:  (kafka.tools.ConsoleConsumer$)
    org.apache.kafka.common.errors.TimeoutException
    Processed a total of 9 messages



```python

```
