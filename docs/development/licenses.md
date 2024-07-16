# 3rd part license generation

We use the following tools and processes for generationg 3rd party licenses.

## Go Modules

### Install tools

For Go modules we use a fork of Kubeflow's testing repo with some fixes.

Clone and update to [this branch](https://github.com/SeldonIO/kubeflow-testing/tree/seldon-update).

```
git clone https://github.com/SeldonIO/kubeflow-testing -b seldon-update
```

Go to go-license-tools folder:

```
cd kubeflow-testing/py/kubeflow/testing/go-license-tools/
```

Install

```
python setup.py install
```

### Generation

Run `make update-3rd-party-licenses` in top level folder.