## Standalone Test

There are two versions

 * create_standalone.sh : create clusterwide operator
 * create_standalone_namespace.sh : create namespaced operator in default namespace

Each creates everything that would be created by OLM and runs the Operator deployment as last step.

## Clusterwide

```
./create_standalone.sh
```

Run some models then:

```
delete_standalone.sh
```

## Namespaced

```
./create_standalone_namespace.sh
```

Run some models then:

```
delete_standalone.sh
```

