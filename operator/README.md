# Kubebuilder v2 Seldon Operator

## Development

### Issues

 * Generated CRD is not structural: https://github.com/kubernetes-sigs/controller-tools/issues/304
   There is a PR for this: https://github.com/kubernetes-sigs/controller-tools/pull/312
 

### Testing

If you installed kubebuilder outside of `/usr/local/kubebuilder` then you will need to set the env var `KUBEBUILDER_ASSETS` for example:

```
export KUBEBUILDER_ASSETS=/home/clive/tools/kubebuilder_2.0.0_linux_amd64/bin
```


Start a kind cluster

```
kind create cluster
```

Install CRD and cert-manager

```
make install
make install-cert-manager 
```

Build image

```
make kind-docker-build
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

Now delete the cluster Deployment as we will run the manager locally. After which you can run locally:

```
make run
```

You should now be able to create SeldonDeployments and Webhook calls will hit the local running manager. The same applies if you debug from GoLand. Though for GoLand you will need to export the KUBECONFIG to the debug configuration.

You should delete the Operator running in the cluster at this point.


