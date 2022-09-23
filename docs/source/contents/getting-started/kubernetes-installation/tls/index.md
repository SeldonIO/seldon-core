# TLS

````{note}
This is a work in progress.
````

Seldon can be run with secure control plane and data plane operations. There are three areas of concern:

 * [Control Plane](#control-plane)
 * [Kafka](#kafka)
 * [Data Plane](#data-plane) 

## Control Plane

Kubernetes secrets and mounted files can be used to provide the certificates in PEM format. These are controlled by environment variables of the form:

 * CONTROL_PLANE_TLS_SECRET_NAME: (optional) the name of the namespaced secret which holds the certificates
 * CONTROL_PLANE_TLS_KEY_LOCATION: the path to the TLS private key
 * CONTROL_PLANE_TLS_CRT_LOCATION: the path to the TLS private key
 * CONTROL_PLANE_TLS_CA_LOCATION: the path to the TLS private key 

Certificates will be loaded and used for the control plane gRPC services. The secrets or folders will be watched for updates (on certificate renewal) and automatically loaded again.

### Helm Control Plane Install

When installing `seldon-core-v2-setup` you can set the secret names for your certificates. If using cert-manager example discussed below this would be (using `seldon-mesh` as the example namespace)::

```bash
helm install seldon-v2 k8s/helm-charts/seldon-core-v2-setup/ -n seldon-mesh \
       --set security.controlplane.protocol=SSL \
       --set security.controlplane.ssl.server.secret=seldon-controlplane-server \
       --set security.controlplane.ssl.client.secret=seldon-controlplane-client
```

## Kafka

You can ensure the components that talk to Kafka run with mTLS. 

 * [Strimzi example](strimzi.md)
 * [AWS MSK example](msk.md)

## Data Plane

````{note}
WIP
````

## Cetificate Providers

The installer/cluster controller for Seldon needs to provide the certificates. As part of Seldon we provide an example set of certificate issuers and certificates using [cert-manager](https://cert-manager.io/).

From the project root run:

### Raw YAML

Raw yaml Certificates can be created with:

```
kubectl create -f k8s/yaml/certs.yaml -n seldon-mesh
```

### Helm

You can install Certificates into the desired namespace, here we use `seldon-mesh` as an example.

```
helm install seldon-v2-certs k8s/helm-charts/seldon-core-v2-certs/ -n seldon-mesh
```

```{toctree}
:maxdepth: 1
:hidden:

strimzi.md
msk.md
```
