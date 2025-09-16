---
description: Learn how to integrate Seldon Core 2 with Schema Registry on both Confluent Cloud and Confluent Platform for having a centralized repository for schema management
---

# Schema Registry

[Schema Registry](https://docs.confluent.io/platform/current/schema-registry/index.html) allows producers and consumers to ensure data consistency and compatibilty 
checks as schemas evolve. It also allows integration of further Confluent Cloud solutions such as connectors.

## Installation
We recommend to use managed Schema registry for production installation. This allows to take away all the complexity on 
running a secure schema registry cluster. 

We currently have tested integration with the following Schema registry solutions:
- Confluent Cloud (security: SASL/PLAIN)
- Self-hosted (security: SASL/PLAIN)

# Helm installation
To install and enable Schema Registry in your cluster with helm provide a configuration file with the following values filled in:
```yaml
security:
  schemaRegistry:
    client:
      URL: (url or Public endpoint)
      username: (api key)
      password: (api key secret)
```

Installing using a values yaml file
```shell
helm upgrade seldon-core-v2-setup seldon-charts/seldon-core-v2-setup \
 --namespace seldon-mesh --set controller.clusterwide=true \
 -f myvalues.yaml \
 --install
```

set custom values 
```shell
helm upgrade seldon-core-v2-setup seldon-charts/seldon-core-v2-setup \
 --namespace seldon-mesh --set controller.clusterwide=true \
 --set security.schemaRegistry.client.URL=schema_registry_url_endpoint \
 --set security.schemaRegistry.client.username=schema_registry_username \
 --set security.schemaRegistry.client.password=schema_registry_password \
 --install
```

## Ansible
We provide automation around the installation of a Kafka cluster for Seldon Core 2 to help with
development and testing use cases.
You can follow the steps defined [here](../../getting-started/kubernetes-installation/ansible.md) to
install Kafka via ansible.