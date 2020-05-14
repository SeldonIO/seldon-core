## Test Kafka Cluster using Strimzi

## Smoke test

 * Start kind cluster
 * install, create cluster and topic using makefile
 * assuming cluster is running in default: get nodeport

```
kubectl get service my-cluster-kafka-external-bootstrap -n default -o=jsonpath='{.spec.ports[0].nodePort}{"\n"}'
```

Use nodeport in consumer, producer on default Kind ip address, e.g.

```
kafka-console-consumer.sh --bootstrap-server 172.17.0.2:31415 --topic my-topic --from-beginning
```

```
kafka-console-producer.sh --broker-list 172.17.0.2:31415 --topic my-topic
```

