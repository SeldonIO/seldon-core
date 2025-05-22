# Scaling in Seldon Core 2

Seldon Core 2 provides multiple approaches to scaling your machine learning deployments, allowing you to optimize resource utilization and handle varying workloads efficiently. This document provides an overview of the different scaling mechanisms available.

## Overview of Scaling Approaches

Seldon Core 2 offers three main approaches to scaling:

1. **Manual Scaling**: Direct control over replica counts for models, servers, and Seldon services
2. **HPA for Models with Server Autoscaling**: Configure scaling of Models using HPA, and Seldon will handle the autoscaling of server replicas
3. **Model Autoscaling with HPA for Servers**: Dynamic scaling of model replicas based on inference load (non-configurable scaling logic)

## Manual Scaling

Manual scaling provides the most direct control over your deployment's resources. You can set specific replica counts for:

- **Models**: Control the number of model replicas through the `replicas` field in the Model CR
- **Servers**: Manage server replicas through the `replicas` field in the Server CR
- **Internal Components**: Various control and dataplane components can be scaled based on your needs

This approach is useful when you have predictable workloads or specific resource requirements.

## HPA Scaling of Models with Server Autoscaling

Server autoscaling is particularly important for Multi-Model Serving (MMS) deployments. It automatically adjusts the number of server replicas based on model requirements, even if there are multiple Models on shared Servers. **Key Features**:
- Automatically scales server replicas in response to model replica changes
- Supports both scale-up and scale-down operations
- Includes policies for empty server removal and lightly loaded server consolidation
- Ensures efficient resource utilization through model packing

## Model Autoscaling with HPA for Servers

Model autoscaling dynamically adjusts the number of model replicas based on inference load. **Key Features**:
- Scales models based on inference lag and inactivity. This cannot be configured.
- Works within defined minimum and maximum replica bounds
- Integrates with server autoscaling for complete resource management
- Supports memory overcommit for efficient resource utilization

## Configuration

Each scaling approach can be configured through:
- Model and Server Custom Resources
- Environment variables
- Helm chart values during installation

For detailed configuration options, refer to the specific documentation for each scaling approach.

## Limitations and Considerations

- Users should be careful about the interaction of multiple approaches. For example when `minReplicas` and `maxReplicas` are set in Model CRDs *and* in HPA manifests targetting those models, the autoscaling will not work as expected.
- Server scaling down may not remove specific replicas due to StatefulSet behavior
- Model autoscaling requires careful threshold configuration
- Server packing policies are experimental and should be tested before production use

For more detailed information about each scaling approach, refer to the specific documentation:
- [Manual Scaling](./manual-scaling.md)
- [Server Autoscaling](./server-autoscaling.md)
- [Model Autoscaling](./autoscaling.md)
