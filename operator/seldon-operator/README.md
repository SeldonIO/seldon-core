# Seldon Operator for RedHat

## Resource Generation

For a new release:

Regenerate the core yaml from kustomize for the "lite" version of Core:

```
make generate-resources
```

Recreate the core yaml from these resources:

```
make deploy/operator.yaml
```

Create a new rule in the Makefile to generate the operator CSV from a previous version using the latest yaml. For 1.1.0 this is an initial rule based off a phony previous release 1.0.0.

```

```
