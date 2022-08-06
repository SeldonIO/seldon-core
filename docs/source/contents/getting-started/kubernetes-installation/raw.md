# Raw Yaml Install

We provide raw yaml to install the CRDs and components.

## Install

From the project root run:

```
make deploy-k8s
```

To install in a particular namespace set the environment variable SELDON_NAMESPACE, e.g.

```bash
export SELDON_NAMESPACE=test
make deploy-k8s
```

## Uninstall

From the project root run:

```
make undeploy-k8s
```

## Raw YAML

If you wish to run the yaml yourself it can be found in `./k8s/yaml`. The steps used by the project Makefile are:

```{literalinclude} ../../../../../Makefile
:language: shell
:start-after: Start raw deploy
:end-before: End raw deploy
```

To uninstall:

```{literalinclude} ../../../../../Makefile
:language: shell
:start-after: Start raw undeploy
:end-before: End raw undeploy
```



