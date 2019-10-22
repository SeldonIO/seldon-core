# End to End Tests

We use [Kind](https://github.com/kubernetes-sigs/kind) to run our tests.

To get everything setup run:

```
kind_test_setup.sh
```

Activate kind kubernetes config:

```
export KUBECONFIG="$(kind get kubeconfig-path)"
```

Then to run the tests:

```
make test
```


