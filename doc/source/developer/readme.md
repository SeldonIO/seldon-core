# Developer

We welcome new contributors. Please read the [code of conduct](https://github.com/SeldonIO/seldon-core/blob/master/CODE_OF_CONDUCT.md) and [contributing guidelines](https://github.com/SeldonIO/seldon-core/blob/master/CONTRIBUTING.md)

## Operator Development

The Operator which manages the SeldonDeployment CRD is contained within the `/operator` folder. It is created using [kubebuilder](https://book.kubebuilder.io/)

For local development we use [kind](https://kind.sigs.k8s.io/), create a kind cluster

```
kind create cluster
```

Install cert-manager

```
make install-cert-manager
```

To build and load the current controller image into the Kind cluster:

```
make kind-image-install
```

To install the Operator run:

```
make deploy
```

If you wish to install the Operator and prepare the controller to be run outside the cluster, for example inside an IDE (we use GoLand) then run:

```
make deploy-local
```




## Tools we use

 - [github-changelog-generator](https://github.com/skywinder/github-changelog-generator)
 - [Grip - Local Markdown viewer](https://github.com/joeyespo/grip)

## Building Seldon Core

* [Build using private repository](build-using-private-repo.md)

## Seldon Prow

 - [prow status](https://prow.seldon.io)

