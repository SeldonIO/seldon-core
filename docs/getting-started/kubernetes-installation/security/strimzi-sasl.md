# Strimzi SASL Example

Create a Strimzi Kafka cluster with SASL_SSL enabled.
This can be done with our Ansible scripts by running the following from the `ansible/` folder:

```
ansible-playbook playbooks/setup-ecosystem.yaml -e @../k8s/samples/ansible-strimzi-kafka-sasl-scram.yaml -e strimzi_kafka_operator_feature_gates=""
```

The referenced SASL/SCRAM YAML file looks like the below:
```{literalinclude} ../../../../../../k8s/samples/ansible-strimzi-kafka-sasl-scram.yaml
:language: yaml
```

This will use the Strimzi Helm chart provided in Core v2.
This will call the Strimzi cluster Helm chart provided by the project with overrides for the cluster authentication type and will also create a user `seldon` with password credentials in a Kubernetes Secret.

Install Core v2 with SASL settings using a custom values file.
This sets the secret created by Strimzi for the user created above (`seldon`) and targets the server certificate authority secret from the name of the cluster created on install of the Kafka cluster (`seldon-cluster-ca-cert`).

Configure Seldon Core v2 by setting following Helm values:

```{literalinclude} ../../../../../../k8s/samples/values-strimzi-kafka-sasl-scram.yaml
:language: yaml
```

```
helm install seldon-v2 k8s/helm-charts/seldon-core-v2-setup/ -n seldon-mesh -f k8s/samples/values-strimzi-kafka-sasl-scram.yaml
```
