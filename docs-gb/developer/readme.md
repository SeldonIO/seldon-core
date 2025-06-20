# Developer

We welcome new contributors.
Please read the [code of
conduct](https://github.com/SeldonIO/seldon-core/blob/master/CODE_OF_CONDUCT.md)
and the [contributing guidelines](contributing.rst).

## Operator Development

The Operator which manages the SeldonDeployment CRD is contained within the `/operator` folder. It is created using [kubebuilder](https://book.kubebuilder.io/)

For local development we use [kind](https://kind.sigs.k8s.io/), create a kind cluster

```console
kind create cluster
```

Install cert-manager

```console
make install-cert-manager
```

To build and load the current controller image into the Kind cluster:

```console
make kind-image-install
```

To install the Operator run:

```console
make deploy
```

If you wish to install the Operator and prepare the controller to be run outside the cluster, for example inside an IDE (we use GoLand) then run the following. This has only been tested on a local Kind cluster:

```console
make deploy-local
```

When the everything is running delete the seldon-controller-manager deployment from the seldon-system namespace as we will run locally.

Next, download the webhook certificate (created by cert-manager) locally:

```console
make tls-extract
```

You can now run the manager locally. You will need to set the webhook-port on startup e.g.,

```console
go run ./main.go --webhook-port=9000
```

If running inside an IDE and you are using Kind then make sure you set the KUBECONFIG env as well.

## Tools we use

- [github-changelog-generator](https://github.com/skywinder/github-changelog-generator)
- [Grip - Local Markdown viewer](https://github.com/joeyespo/grip)

## Building Seldon Core

- [Build using private repository](build-using-private-repo.md)

