# Kafka Integration

## Install Kafka

Clone https://github.com/SeldonIO/ansible-k8s-collection

Create a kafka operator install in kafka namespace

```
ansible-playbook playbooks/kafka.yaml
```

Create our Kafka cluster

```
kubectl create -f cluster.yaml -n kafka
```

