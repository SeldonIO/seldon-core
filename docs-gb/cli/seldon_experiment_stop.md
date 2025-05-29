---
description: Terminate and halt Seldon Core experiments using the seldon experiment stop CLI command. This command allows you to safely stop running experiments, manage control plane operations, and clean up experiment resources in your Seldon Core deployment.
---

## seldon experiment stop

stop an experiment

### Synopsis

stop an experiment

```
seldon experiment stop <experimentName> [flags]
```

### Options

```
      --authority string        authority (HTTP/2) or virtual host (HTTP/1)
  -f, --file-path string        experiment manifest file (YAML)
      --force                   force control plane mode (load model, etc.)
  -h, --help                    help for stop
      --scheduler-host string   seldon scheduler host (default "0.0.0.0:9004")
  -v, --verbose                 verbose output
```

### SEE ALSO

* [seldon experiment](seldon_experiment.md)	 - manage experiments

