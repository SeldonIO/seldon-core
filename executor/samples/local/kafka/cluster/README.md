# Install Kafka

```
make install
```

Wait for Strimzi operator pod to start

```
make create_cluster
```

Wait for zookeeper and cluster pods to start


```
make create_topic
```

## Smoke test

Start a consumer

```
make start_consumer
```

Start a producer

```
make start_producer
```

enter messages on command line and check they are received.

