# Strimzi mTLS Example

## Cluster Setup

If you have installed Strimzi we have an example Helm chart to create a Kafka cluster for seldon and an associated user in `kafka/strimzi` folder. Ensure the `tls` is enabled with:

```
broker:
  tls:
    enabled: true
    port: 9093
    listenerType: internal
    authentication:
      type: tls
```

The Ansible `setup-ecosystem` playbook will also install Strimzi and include a mTLS endpoint. See [here](../ansible.md).

## mTLS Example

Create a Kafka User `seldon` in the namespace seldon was installed. This assumes Strimzi Kafka cluster is installed in the same namespace or is running with cluster wide permissions. Our Ansible scripts to setup the ecosystem will also create this user if tls is active.

```{literalinclude} ../../../../../../k8s/samples/strimzi-example-tls-user.yaml
:language: yaml
```

If you don't have this user you can install it with in your desired namespace (here `seldon-mesh`):

```
kubectl create -f k8s/samples/strimzi-example-tls-user.yaml -n seldon-mesh
```

Install seldon with the Strimzi certificate secrets using a custom values file. This sets the secret created by Strimzi for the user created above (`seldon`) and targets the server certificate authority secret from the name of the cluster created on install of the Kafka cluster (`seldon-cluster-ca-cert`).

Configure Seldon Core v2 by setting following Helm values:

```{literalinclude} ../../../../../../k8s/samples/values-strimzi-kafka-mtls.yaml
:language: yaml
```


```
helm install seldon-v2 k8s/helm-charts/seldon-core-v2-setup/ -n seldon-mesh -f k8s/samples/values-strimzi-kafka-mtls.yaml
```

You can now go ahead and install a SeldonRuntime in your desired install namespace (here `seldon-mesh`), e.g.

```
helm install seldon-v2-runtime ../k8s/helm-charts/seldon-core-v2-runtime  -n seldon-mesh
```
