---
description: Learn how to integrate Seldon Core 2 with Schema Registry on both Confluent Cloud and Confluent Platform for having a centralized repository for schema management
---

# Schema Registry

[Schema Registry](https://docs.confluent.io/platform/current/schema-registry/index.html) provides centralized schema management for data consistency and compatibility. Using Schema Registry in Core 2 enables seamless integration with Kafka Connect, ksqlDB, and other Confluent ecosystem components.

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

1. Sets environment variables to the value of `security.schemaRegistry.configPath`:
   - Dataflow: `SELDON_KAFKA_SCHEMA_REGISTRY_CONFIG_PATH`
   - Model Gateway and Pipeline Gateway: `SCHEMA_REGISTRY_CONFIG_PATH`
2. Creates and mounts a volume `kafka-schema-volume` at `/mnt/schema-registry` for Dataflow, Pipeline Gateway, and Model Gateway
3. Mounts the `confluent-schema` secret to the `kafka-schema-volume`
4. Expects a `.confluent-schema.yaml` configuration file as a key in the `confluent-schema` secret

{% hint style="info" %}
**Note**: When using Helm installation, `security.schemaRegistry.configPath` must be set to `/mnt/schema-registry` 
because Helm automatically creates and mounts the volume at this path. For custom installations where you manually 
configure volumes or run outside of Kubernetes, you can set this to any directory path where your 
`.confluent-schema.yaml` file is located.
{% endhint %}

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
