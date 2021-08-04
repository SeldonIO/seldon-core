# Kubebuilder v2 Seldon Operator

## Development

We assume:

 * Go 1.13
 * Kubebuilder 2.3.0
 * golangci-lint v1.35.2

### Issues

 * Generated CRD is not structural: https://github.com/kubernetes-sigs/controller-tools/issues/304
   There is a PR for this: https://github.com/kubernetes-sigs/controller-tools/pull/312

### Prerequisites

For running locally `kind`, `kustomize` and `kubebuilder` should be installed.

For linting the `golangci-lint` binary is required. To install, follow the [official install instructions](https://golangci-lint.run/usage/install/) by running:
```shell
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.35.2
```

### Testing

If you installed kubebuilder outside of `/usr/local/kubebuilder` then you will need to set the env var `KUBEBUILDER_ASSETS` for example:

```
export KUBEBUILDER_ASSETS=/home/clive/tools/kubebuilder_2.3.0_linux_amd64/bin
```


Start a kind cluster

```
kind create cluster
export KUBECONFIG="$(kind get kubeconfig-path --name="kind")"
```

Install CRD and cert-manager

```
make install
make install-cert-manager
```

Build image

```
make docker-build
```

## Standard Deploy

```
make deploy
```

## Local Development Deploy


Deploy with webhook config set to point to host

```
make deploy-local
```

When running update tls certificates locally

```
make tls-extract
```

Now delete the cluster Deployment as we will run the manager locally:

```
kubectl delete deployment -n seldon-system seldon-controller-manager
``` 

Now we can run locally:

```
make run
```

You should now be able to create SeldonDeployments and Webhook calls will hit the local running manager. The same applies if you debug from GoLand. Though for GoLand you will need to export the KUBECONFIG to the debug configuration.

You should delete the Operator running in the cluster at this point.

# Build Helm Chart

Use the Makefile in the `./helm` directory. Ensure you have `pyyaml` in your python environment.


# Openshift

see [here](openshift.md)