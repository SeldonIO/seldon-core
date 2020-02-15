
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

TODO

