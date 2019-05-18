# Upgrading Notes

## Upgrading to 0.2.8 from previous versions.

### Helm

The helm charts to install Seldon Core have changed. There is now a single Helm chart `seldon-core-operator` that installs the CRD and its controller. Ingress options are now separate and you need to choose between the available options which are at present:

 * Ambassador - via its official Helm chart
 * Seldon's OAUTH Gateway - via its standalone Helm chart

For more details see the [install docs](../workflow/install.md).

The Helm chart `seldon-core-operator` will require clusterwide RBAC and should be installed by a cluster admin.

### Ksonnet

Ksonnet is now deprecated. You should convert to using Helm to install Seldon Core.
