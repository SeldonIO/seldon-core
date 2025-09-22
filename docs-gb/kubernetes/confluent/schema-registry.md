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


## Configuration Details

### Service Integration

When Schema Registry is configured, the following Seldon Core 2 services automatically integrate with it:

- **Dataflow**: Handles data processing workflows
- **Pipeline Gateway**: Manages pipeline inference requests
- **Model Gateway**: Routes model inference traffic

### Environment Configuration

Setting `security.schemaRegistry.configPath` in the Helm values.yaml file configures the services as follows:

1. Sets the `SELDON_KAFKA_SCHEMA_REGISTRY_CONFIG_PATH` environment variable
2. Mounts the `confluent-schema` secret to `/mnt/schema-registry`
3. Expects a `.confluent-schema.yaml` configuration file in the mounted directory

### Configuration File Format

The `.confluent-schema.yaml` file must follow this structure:

```yaml
schemaRegistry:
  client:
    URL: your-schema-registry-endpoint
    username: api-key
    password: api-secret
```

## Subject Registration

Schema subjects are automatically registered when messages are first published to Kafka topics. This occurs during the initial inference request processing by any of the integrated services.

### Subject Naming Strategy

Seldon Core 2 uses the **topic name strategy** for Schema Registry subject naming:

- Subject names are derived directly from Kafka topic names
- Each model automatically creates subjects for both input and output topics
- This ensures consistent schema management across the entire inference pipeline

For more information, see the [Confluent Schema Registry documentation](https://docs.confluent.io/platform/current/schema-registry/fundamentals/serdes-develop/index.html#subject-name-strategy).
