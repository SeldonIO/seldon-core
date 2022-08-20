# TLS

This is a work in progress.

Seldon can be run with secure control plane and data plane operations. There are three areas of concern:

 * [Control Plane](#control-plane)
 * [Data Plane](#data-plane)
 * [Kafka](#kafka)

Kubernetes secrets and mounted files can be used to provide the certificates in PEM format. These are controlled by environment variables of the form:

 * <service>_TLS_SECRET_NAME: the name of the namespaced secret whihc holds the certificates
 * <service>_TLS_FOLDER_PATH: the path to a folder containing the certificates

If both of the above environment variables is empty or absent then plain text service connections will be used.
At present, the Kubernetes Secret keys or the filesystem folder filenames need to follow the following format:

 * tls.key : private key certificate
 * tld.crt : public key certificate
 * ca.crt : certificate authority certificate

Certificates will be loaded and used for the desired gRPC services. The secrets or folders will be watched for updates (on certificate renewal) and automatically loaded again.

## Control Plane

For the control plane mTLS can be used on gRPC services:

 * Scheduler gRPC service: envar service name: SCHEDULER
 * Agent gRPC service: WIP
 * Dataflow gRPC service: WIP

### Helm Control Plane Install

When installing `seldon-core-v2-setup` you can set the secret names for your certificates. If using cert-manager example discussed below this would be (using `seldon-mesh` as the example namespace)::

```bash
helm install seldon-v2 k8s/helm-charts/seldon-core-v2-setup/ -n seldon-mesh \
     --set scheduler.tls.scheduler.server.secret=seldon-scheduler-server \
     --set scheduler.tls.scheduler.client.secret=seldon-scheduler-client
```

## Data Plane

WIP

## Kafka

WIP


## Cetificate Providers

The installer/cluster controller for Seldon needs to provide the certificates. As part of Seldon we provide an example set of certificate issuers and certificates using [cert-manager](https://cert-manager.io/).

From the project root run:

### Raw YAML

```
kubectl create -f k8s/yaml/certs.yaml
```

### Helm

You can install into the desired namespace, here we use `seldon-mesh` as an example.

```
helm install seldon-v2-certs k8s/helm-charts/seldon-core-v2-certs/ -n seldon-mesh
```