# Kafka

To take advantage of Pipelines and the full dataflow architecture you need to have a Kafka cluster installed that Seldon can use.

We list alternatives below:

## Strimzi

[Strimzi](https://strimzi.io/)  provides a simple Kubernetes operator installation of Kafka. To install via Ansible:

  * Clone https://github.com/SeldonIO/ansible-k8s-collection
  * Create a kafka operator install in kafka namespace
  ```bash
  ansible-playbook playbooks/kafka.yaml
  ```
  * Create our Kafka cluster by installing `kafka/strimzi/cluster.yaml`
  ```
  kubectl create -f cluster.yaml -n kafka
  ```


