# Tools to create Ksonnet core JSON

The core Ksonnet JSON is taken from the Helm charts to ensure lock-step with the helm releases.

To create the latest json run

```
make build
```

To copy to ksonnet folder

```
make release
```

We support kubernetes API >= 1.8.6 at present. There are Makefile targets to get kubectl and start and v1.8.6 minikube cluster.



