# End to End Tests

## Prerequisites
We use [Kind](https://github.com/kubernetes-sigs/kind), [S2I](https://github.com/openshift/source-to-image) and [Kustomize](https://github.com/kubernetes-sigs/kustomize/blob/master/docs/INSTALL.md).

Install ```python -m pip install grpcio-tools```

## Setup
To get everything setup run:

```
kind_test_setup.sh
```

Activate kind kubernetes config:

```
export KUBECONFIG="$(kind get kubeconfig-path)"
```

## Run
Then to run the tests and log output to a file:

```
make test > test.log
```

## Logs
To view test logs in a separate terminal:
```
tail -f test.log
```

To also follow controller logs in a separate terminal:
```
export KUBECONFIG="$(kind get kubeconfig-path)"
kubectl logs -f -n seldon-system $(kubectl get pods -n seldon-system -l app=seldon -o jsonpath='{.items[0].metadata.name}')
```
