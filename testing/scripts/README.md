# End to End Tests

## Running the test suite

### Prerequisites

We use:

[Kind](https://github.com/kubernetes-sigs/kind) v0.6.1

[S2I](https://github.com/openshift/source-to-image) v1.1.14

[Kustomize](https://github.com/kubernetes-sigs/kustomize/blob/master/docs/INSTALL.md) v3.2.3

[Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) v1.16.2

[Pytest](https://docs.pytest.org/en/latest/getting-started.html#install-pytest)
`pip install -U pytest`

[gRPC.io Tools](https://grpc.io/docs/languages/python/quickstart/#grpc-tools)
`python -m pip install grpcio-tools`

### Setup

To get everything setup run:

```console
kind_test_setup.sh
```

Activate kind kubernetes config:

```console
export KUBECONFIG="$(kind get kubeconfig-path)"
```

### Run

Then to run the tests and log output to a file:

```console
make test_sequential > test_sequential.log
```
```console
make test_parallel > test_parallel.log
```
```console
make test_notebooks > test_notebooks.log
```

To run a single test:
```console
pytest -sv test_tags.py::TestTagsPythonS2iK8s::test_model_single_grpc > test_model_single_grpc.log
```

### Logs

To view test logs in a separate terminal:

```console
tail -f test_sequential.log
```

To also follow controller logs in a separate terminal:

```console
export KUBECONFIG="$(kind get kubeconfig-path)"
kubectl logs -f -n seldon-system $(kubectl get pods -n seldon-system -l app=seldon -o jsonpath='{.items[0].metadata.name}')
```

## Writing new tests

### Invididual namespaces

Each test should run on its own separate Kubernetes namespace.
This allows for a cleaner environment as well as enables some of them
to get parallelised (see [Sequential and parallel
tests](#Sequential-and-parallel-tests)).

To make the creation and deletion of the namespace easier, you can
use the `namespace` fixture.
This fixture takes care of creating a namespace with a unique name
and tearing it down at the end.
To use it you just need to specify the `namespace` parameter as part
of the arguments passed to your test function.
The fixture will create the new namespace on the background and the
`namespace` argument will take on the namespace name as value.

```python
def test_foo(..., namespace, ...):
  print(f"the new namespace's name is {namespace}")
```

### Sequential and parallel tests

We leverage `pytest-xdist` to speed up the test suite by
parallelising the test execution.
However, some of the tests may have side-effects which make
them unsuitable to be executed alongside others.
For example, the operator update tests change the cluster-wide Seldon
operator which could affect the other tests running in parallel.

To differentiate between parallelisable tests and tests which are
required to be run sequentially you can use the `sequential` mark.
Marks in `pytest` allow to [select subsets of
tests](http://doc.pytest.org/en/latest/example/markers.html).

To mark a test to run sequentially, you can do:

```python
@pytest.mark.sequential
def test_foo(...):
  print("the scripts will run this test sequentially")
```

The integration test scripts will make sure that tests marked as
`sequential` get run with a single worker.
