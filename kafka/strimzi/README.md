# Kafka Integration

## Install Kafka

Clone https://github.com/SeldonIO/ansible-k8s-collection

Create a kafka operator install in kafka namespace

create playbooks/kafka_scv2.yaml

```
- name: Install Kafka
  hosts: localhost
  roles:
    - strimzi_kafka
  vars:
    strimzi_kafka_create_cluster: false
    strimzi_kafka_create_metrics: false
```

```
ansible-playbook playbooks/kafka_scv2.yaml
```

Create our Kafka cluster

```
kubectl create -f cluster.yaml -n kafka
```

