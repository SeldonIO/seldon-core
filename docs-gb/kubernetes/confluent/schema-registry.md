---
description: Learn how to integrate Seldon Core 2 with Schema Registry on both Confluent Cloud and Confluent Platform for having a centralized repository for schema management
---

# Schema Registry

[Schema Registry](https://docs.confluent.io/platform/current/schema-registry/index.html) provides centralized schema management for data consistency and compatibility.

## Quick Installation

**Prerequisites**: Schema Registry endpoint with SASL/PLAIN authentication (Confluent Cloud or self-hosted).

### Step 1: Create Schema Registry Secret

Replace the placeholder values with your actual credentials:

```bash
kubectl create secret generic confluent-schema --from-literal=.confluent-schema.yaml='
schemaRegistry:
  client:
    URL: your-schema-registry-endpoint
    username: api-key
    password: api-secret'
```

### Step 2: Install with Helm

```bash
helm upgrade seldon-core-v2-setup seldon-charts/seldon-core-v2-setup \
  --namespace seldon-mesh \
  --set security.schemaRegistry.configPath=/mnt/schema-registry \
  --install
```

That's it! The model-gateway, pipeline-gateway, and dataflow services will automatically mount the secret at `/mnt/schema-registry` and use it for Schema Registry authentication.

## Ansible
We provide automation around the installation of a Kafka cluster for Seldon Core 2 to help with
development and testing use cases.
You can follow the steps defined [here](../../getting-started/kubernetes-installation/ansible.md) to
install Kafka via ansible.