
# Seldon Batch Processing Kafka Data Source

This example will walk you through the steps for you to leverage the data ingestor batch functionality in Seldon Core to build your own custom Kafka Data Ingestor which will consume from a Kafka Topic, send the data to Seldon and then publish the output into another Kafka topic. The diagram below provides some intuition on what we'll be deploying:

[](./batch-seldon-kafka-cluster.png)

In this example we will:
1. Create a Kafka Data Ingestor
2. Run our Kafka Data Ingestor with docker-compose
3. Test the Data Ingestor in Kubernetes Cluster with test model

## 1) Create Kafka Data Ingestor

To get started we want to create a simple data ingestor that consumes and publishes from Kafka.

For this, we will create the file `ingestor/KafkaDataIngestor.py` with the following contents:

```python
import kafka
from multiprocessing import get_logger

logger = get_logger()

class KafkaDataIngestor:

    def __init__(
            self,
            bootstrap_servers,
            consumer_topic,
            producer_topic,
            group_id):
        """
        This includes the inputs from the parameters, which we could provide some by default
        """
        self.bootstrap_servers = bootstrap_servers
        self.consumer_topic = consumer_topic
        self.producer_topic = producer_topic
        self.group_id = group_id

    def connection(self):
        consumer = kafka.KafkaConsumer(
            self.consumer_topic,
            group_id=self.group_id,
            bootstrap_servers=self.bootstrap_servers)
        producer = kafka.KafkaProducer(
            bootstrap_servers=self.bootstrap_servers)
        return {
            "consumer": consumer,
            "producer": producer,
            "producer_topic": self.producer_topic
        }

    @staticmethod
    def fetch(self, connection):
        """
        Callects the single next message in the queue to process or
        waits until there is a new message. It leverages the yield
        functionality that the KafkaConsumer exposes.
        """
        return next(connection["consumer"])

    @staticmethod
    def process(seldon_client, connection, in_data):
        out_data = seldon_client.predict(data=in_data)
        return out_data

    @staticmethod
    def publish(out_data, in_data, connection):
        logger.info(f"Publishing with out_data=[{out_data}], in_data={in_data}")

        res = connection["producer"].send(
            connection["producer_topic"], in_data)

        logger.info(f"Publish result: {res}")

```

## 2) Containerise the data ingestor 

In order to containerise the data ingestor we need to create a new file `ingestor/s2i/environment` with the following contents:

```bash
MODEL_NAME=KafkaDataIngestor
API_TYPE=BATCH
```

With this now we just have to containerise our wrapper using the Seldon CLI tools by running:

```
s2i build ingestor/. seldonio/seldon-core-s2i-python37:0.18-SNAPSHOT kafka_data_ingestor:0.1
```

## 3. Test the Data Ingestor in Kubernetes Cluster with test model

We now want to test our data ingestor with the existing Seldon Core model samples.

For this you need to make sure you set up a Kubernetes cluster with all Seldon Core dependencies installed [(Operator, ingress, etc)](https://docs.seldon.io/projects/seldon-core/en/latest/workflow/install.html).

Once you have everything installed, we'll do the following steps:

1. Run Kafka in our Kubernetes Cluster
2. Create Seldon Deployment Config 
3. Deploy our Seldon Deployment to our cluster
4. Publish messages and see messages coming out from kafka topics

### 1. Run Kafka in our Kubernetes Cluster

We first need to make sure our helm installer has access to the incubator charts:

```console
helm repo add bitnami https://charts.bitnami.com/bitnami 
```

Now we're able to cre
ate a simple Kafka deployment:

```console
helm install my-kafka bitnami/kafka
```

Once it's running we'll be able to see the containers:

```console
$ kubeclt get pods

NAME                   READY   STATUS    RESTARTS   AGE
my-kafka-0             1/1     Running   0          2m43s
my-kafka-1             1/1     Running   0          42s
my-kafka-zookeeper-0   1/1     Running   0          2m43s
my-kafka-zookeeper-1   1/1     Running   0          96s
my-kafka-zookeeper-2   1/1     Running   0          62s
```


### 2) Create a Seldon Deployment that uses our model

Now we want to create a Seldon Deploymen configuration file that we'll be able to deploy.

For this we'll use a simple file that just sets the following:
* Selects the deployment to run without the Engine/Orchestrator
* Adds the environment variables to point it to the cluster we just deployed
* Points to the docker image that we just built

The contents of `cluster/kafka_batch_deployment.yaml` are as follows:

```yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: iris-streaming-deployment
spec:
  name: iris-streaming-spec
  predictors:
  - name: iris-streaming-predictor
    batch:
      streaming: 'true'
      name: kafka-ingestor
      parameters:
      - name: bootstrap_servers
        type: STRING
        value: kafka://my-kafka.default.svc.cluster.local:9092
      - name: consumer_topic
        type: STRING
        value: iris-streaming-deployment-kafka-ingestor-input
      - name: producer_topic
        type: STRING
        value: iris-streaming-deployment-kafka-ingestor-output
      - name: group_id
        type: STRING
        value: iris-streaming-deployment-kafka-ingestor
    graph:
      name: iris-model
      endpoint:
        type: REST
      type: MODEL
      children: []
      parameters: []
    componentSpecs:
    - spec:
        containers:
        - image: seldonio/sklearn-iris:0.1
          name: iris-model
        - image: kafka_data_ingestor:0.1
          name: kafka-ingestor
```


### 3) Deploy our Seldon Deployment

Now that we've created out deployment, we just need to launch it:

```console
kubectl apply -f cluster/streaming_model_deployment.yaml
```

Once it's deployed we can see it by running:

```console
$ kubectl get pods | grep streaming

streaming-spec-streaming-graph-e90bdcd-56986c5d4b-7xvtm   1/1     Running   0          6m28s
```

4) Publish messages in the input topic and see messages coming from output topic

Now we want to test it by sending some messages.

We can get the name of the pod by running:

```console
export STREAM_SELDON_POD=`kubectl get pod -l seldon-app=streaming-deployment-streaming-spec-streaming-graph -o jsonpath="{.items[0].metadata.name}"`
```

First let's run a consumer to see the output:

```python
kubectl exec -i $STREAM_SELDON_POD python - <<EOF
import kafka
consumer = kafka.KafkaConsumer(
    'streaming-deployment-streaming-model-predict-output',
    bootstrap_servers='my-kafka:9092');
print(next(consumer).value)
EOF
```

Then let's send the message:

```python
kubectl exec -i $STREAM_SELDON_POD python - <<EOF
import kafka, json;
producer = kafka.KafkaProducer(bootstrap_servers='my-kafka:9092', value_serializer=lambda v: json.dumps(v).encode('utf-8'));
result = producer.send('streaming-deployment-streaming-model-predict-input', value={'data': { 'ndarray': [1,2,3,4] } })
result.get(timeout=3)
EOF
```

We can now see in our consumer that we have received and printed the output as follows:

```json
b'{"data": {"ndarray": [1, 2, 3, 4], "names": []}, "meta": {}}'
```




