# Strimzi SASL Example

Create a Strimzi Kafka cluster with SASL_SSL enabled. This can be done with our Ansible scripts during ecosystem setup by running from the project ansible folder:

```
ansible-playbook playbooks/setup-ecosystem.yaml -e kafka_cluster_values_files=${PWD}/../k8s/samples/ansible-strimzi-kafka-sasl-scram.yaml -e strimzi_kafka_operator_feature_gates=""
```

This will call the Strimzi cluster Helm chart provided by the project with overrides for the cluster authentication type and will also create a user `seldon` with password credentials in a Kubernetes Secret:

```{literalinclude} ../../../../../../k8s/samples/ansible-strimzi-kafka-sasl-scram.yaml
:language: yaml
```

Install seldon with sasl settings using a custom values file. This sets the secret created by Strimzi for the user created above (`seldon`) and targets the server certificate authority secret from the name of the cluster created on install of the Kafka cluster (`seldon-cluster-ca-cert`). 

```{literalinclude} ../../../../../../k8s/samples/values-strimzi-kafka-sasl-scram.yaml
:language: yaml
```

```
helm install seldon-v2 k8s/helm-charts/seldon-core-v2-setup/ -n seldon-mesh -f k8s/samples/values-strimzi-kafka-sasl-scram.yaml
```
