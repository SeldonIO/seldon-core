
# Seldon Async Stream Processing

This example will walk you through the steps for you to be able to create a model and deploy it as a streaming service, as Seldon supports REST, GRPC and KAFKA servers. 

In this example we will:
1) Create a simple model wrapper
2) Containerise the model
3) Test the model locally
4) Deploy the model and test in a kubernetes cluster

## 1) Create  a simple model wrapper

To get started we want to create a simple model wrapper. 

For this we will create the file `model/StreamingModel.py` with the following contents:

```
class StreamingModel:
    def __init__(self):
        print("INITIALIZING STREAMINGMODEL")

    def predict(self, data, names=[], meta=[]):
        print(f"Inside predict: data [{data}] names [{names}] meta [{meta}]")
        return data
```

Key thing to note is that Streaming models are limited to only the predict function for now.

This means that Routes and Aggregators are still in development.

## 2) Containerise the model

In order to containerise the model we need to create a new file `model/.s2i/environment` with the following contents:

```
MODEL_NAME=StreamingModel
API_TYPE=KAFKA
SERVICE_TYPE=MODEL
PERSISTENCE=0
```

With this now we just have to containerise our wrapper using the Seldon CLI tools by running:

```
s2i build model/. seldonio/seldon-core-s2i-python37:0.17-SNAPSHOT streaming_model:0.1
```

## 3) Test the model locally

In order to test the model locally we need to perform the following steps:

3.1) Set up the container env variables
3.2) Run a kafka cluster locally so that our model can connect to it
3.3) Run the container to listen to data in kafka input topic
3.3) Send messages to the Kafka input topic
3.4) Listen for messages on the output topic 

### 3.1) Provide environment variables

When our model runs in kubernetes managed by Seldon, it's provided with key environment variables.

Given we'll be testing it locally first, we will have to add these environment variables manually.

The easiest way to provide these environment variables is through a file `local/container.env`.

The contents should be the location of the kafka cluster, together with the name of your kubernetes components.


`local/conatiner.env` contents:
```
PREDICTIVE_UNIT_SERVICE_HOST=kafka
PREDICTIVE_UNIT_SERVICE_PORT=9092
PREDICTIVE_UNIT_ID=streaming_model
PREDICTOR_ID=streaming_graph
SELDON_DEPLOYMENT_ID=streaming_deployment
```

Due to our streaming library not supporting opentracing > 2.0.0 we need to add the following temp overrides:
(For more information read https://github.com/robinhood/faust/issues/528)

```
Flask-OpenTracing==0.2.0
jaeger-client==3.13.0
opentracing==1.3.0
```

### 3.2) Run a kafka cluster locally so that our model can connect to it

Now we want to run a kafka cluster. 

For this we will run a simple single-node kafka cluster, and zookeeper worker that will expose the port in localhost.

We will be using docker-compose version 1.25 to run the following `local/local-docker-compose.yaml` file:

```
version: "3"

volumes:
  kafka_zookeeper: {}
  kafka_kafka: {}

networks:
  kafkanet:
    driver: bridge

services:

  zookeeper:
    image: wurstmeister/zookeeper:latest
    container_name: zookeeper
    ports:
      - "2181:2181"
    volumes:
      - kafka_zookeeper:/opt/zookeeper-3.4.13/data
    networks:
      - kafkanet

  kafka:
    image: wurstmeister/kafka:2.12-2.2.0 
    container_name: kafka
    command: [start-kafka.sh]
    ports:
      - "8080:8080"
      - "9092:9092"
    environment:
      KAFKA_ADVERTISED_HOST_NAME: kafka
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_PORT: 9092
    volumes:
      - kafka_kafka:/opt/kafka_2.12-2.2.0/logs
    depends_on:
      - "zookeeper"
    networks:
      - kafkanet

```

We can run this file by simply running the following command:

```
docker-compose -f local/local-docker-compose.yaml up -d
```

This is running deattached, so you can see the logs by running:

```
docker-compose -f local/local-docker-compose.yaml logs -f
```

### 3.3) Run model 

TO run our model locally, we can simply use the following command:

```
docker run --name streaming_model --network kafkanet --env-file local/container.env streaming_model:0.1
```

In order to see if it's working we want to run a consumer that listens for the next message - we'll leverage the running container to run the python command directly in the command line:

```
docker exec -i streaming_model python - <<EOF
import kafka
consumer = kafka.KafkaConsumer(
    'streaming_deployment-streaming_model-predict-output',
    bootstrap_servers='kafka:9092');
print(next(consumer).value)
EOF
```

And then we can send some data by also leveraging the running container:

```
docker exec -i streaming_model python - <<EOF
import kafka, json;
producer = kafka.KafkaProducer(bootstrap_servers='kafka:9092', value_serializer=lambda v: json.dumps(v).encode('utf-8'));
result = producer.send('streaming_deployment-streaming_model-predict-input', value={'data': { 'ndarray': [1,2,3,4] } })
result.get(timeout=3)
EOF
```

We can now see that the consumer prints out the content from the output topic:

```
b'{"data": {"ndarray": [1, 2, 3, 4], "names": []}, "meta": {}}'
```


## 4) Deploy the model and test in a Kubernetes cluster

We now want to test it in our Kubernetes cluster. 

For this, you will need to make sure you have all the [Seldon Core dependencies installed (Operator, Ingress, etc).](https://docs.seldon.io/projects/seldon-core/en/latest/workflow/install.html)

Once you have everything installed, we'll do the following steps:
1) Run Kafka in our Kubernetes
2) Create a Seldon Deployment that uses our model
3) Deploy our Seldon Deployment
4) Publish messages in the input topic and see messages coming from output topic

### 1) Run Kafka in our Kubernetes

We first need to make sure our helm installer has access to the incubator charts:

```
helm repo add incubator http://storage.googleapis.com/kubernetes-charts-incubator
```

Now we're able to create a simple Kafka deployment:

```
helm install my-kafka incubator/kafka
```

Once it's running we'll be able to see the containers:

```bash
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

The contents of `cluster/streaming_model_deployment.json` are as follows:

```json
{    "apiVersion": "machinelearning.seldon.io/v1alpha2",
 
   "kind": "SeldonDeployment",
    "metadata": {
        "name": "streaming-deployment",
        "creationTimestamp": null
    },
    "spec": {
        "name": "streaming-spec",
        "predictors": [
            {
                "annotations": {
                    "seldon.io/no-engine": "true"
                },
                "name": "streaming-graph",
                "graph": {
                    "name": "streaming-model",
                    "endpoint": {
                        "type": "REST"
                    },
                    "type": "MODEL",
                    "children": [],
                    "parameters": []
                },
                "componentSpecs": [
                    {
                        "spec": {
                            "containers": [
                                {
                                    "image": "streaming_model:0.1",
                                    "name": "streaming-model",
                                    "env": [
                                        {
                                            "name": "SELDON_LOG_LEVEL",
                                            "value": "DEBUG"
                                        },
                                        {
                                            "name": "PREDICTIVE_UNIT_SERVICE_PORT",
                                            "value": "9092"
                                        },
                                        {
                                            "name": "PREDICTIVE_UNIT_SERVICE_HOST",
                                            "value": "my-kafka"
                                        },
                                        {
                                            "name": "PREDICTIVE_UNIT_ID",
                                            "value": "streaming-model"
                                        },
                                        {
                                            "name": "PREDICTOR_ID",
                                            "value": "streaming-spec"
                                        },
                                        {
                                            "name": "SELDON_DEPLOYMENT_ID",
                                            "value": "streaming-deployment"
                                        }
                                    ]
                                }
                            ],
                            "terminationGracePeriodSeconds": 1
                        }
                    }
                ],
                "replicas": 1,
                "engineResources": {},
                "svcOrchSpec": {},
                "traffic": 100,
                "explainer": {
                    "containerSpec": {
                        "name": "",
                        "resources": {}
                    }
                }
            }
        ],
        "annotations": {
            "seldon.io/engine-seldon-log-messages-externally": "true"
        }
    },
    "status": {}
}
```


### 3) Deploy our Seldon Deployment

Now that we've created out deployment, we just need to launch it:

```
kubectl apply -f cluster/streaming_model_deployment.json
```

Once it's deployed we can see it by running:

```
$ kubectl get pods | grep streaming

streaming-spec-streaming-graph-e90bdcd-56986c5d4b-7xvtm   1/1     Running   0          6m28s
```

4) Publish messages in the input topic and see messages coming from output topic

Now we want to test it by sending some messages.

We can get the name of the pod by running:

```
export STREAM_SELDON_POD=`kubectl get pod -l seldon-app=streaming-deployment-streaming-spec-streaming-graph -o jsonpath="{.items[0].metadata.name}"`
```

First let's run a consumer to see the output:

```
kubectl exec -i $STREAM_SELDON_POD python - <<EOF
import kafka
consumer = kafka.KafkaConsumer(
    'streaming-deployment-streaming-model-predict-output',
    bootstrap_servers='my-kafka:9092');
print(next(consumer).value)
EOF
```

Then let's send the message:

```
kubectl exec -i $STREAM_SELDON_POD python - <<EOF
import kafka, json;
producer = kafka.KafkaProducer(bootstrap_servers='my-kafka:9092', value_serializer=lambda v: json.dumps(v).encode('utf-8'));
result = producer.send('streaming-deployment-streaming-model-predict-input', value={'data': { 'ndarray': [1,2,3,4] } })
result.get(timeout=3)
EOF
```

We can now see in our consumer that we have received and printed the output as follows:

```
b'{"data": {"ndarray": [1, 2, 3, 4], "names": []}, "meta": {}}'
```



