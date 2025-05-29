---
description: Launch and initialize new Seldon Core experiments using the seldon experiment start CLI command. This command helps you begin new machine learning experiments by loading experiment manifests, managing control plane operations, and configuring experiment parameters in Seldon Core.
---

## seldon experiment start

start an experiment

### Synopsis

start an experiment

```
seldon experiment start [flags]
```

### Options

```
      --authority string        authority (HTTP/2) or virtual host (HTTP/1)
  -f, --file-path string        experiment manifest file (YAML)
  -h, --help                    help for start
      --force                   force control plane mode (load model, etc.)
      --scheduler-host string   seldon scheduler host (default "0.0.0.0:9004")
  -r, --show-request            show request
  -o, --show-response           show response (default true)
```

### SEE ALSO

* [seldon experiment](seldon_experiment.md)	 - manage experiments

