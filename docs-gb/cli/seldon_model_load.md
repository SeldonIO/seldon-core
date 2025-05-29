---
description: Deploy and initialize machine learning models in Seldon Core using the seldon model load CLI command. This command helps you load model manifests, manage model deployment in the control plane, and prepare models for inference in your Seldon Core environment.
---

## seldon model load

load a model

### Synopsis

load a model

```
seldon model load [flags]
```

### Options

```
      --authority string        authority (HTTP/2) or virtual host (HTTP/1)
  -f, --file-path string        model manifest file (YAML)
      --force                   force control plane mode (load model, etc.)
  -h, --help                    help for load
      --scheduler-host string   seldon scheduler host (default "0.0.0.0:9004")
  -r, --show-request            show request
  -o, --show-response           show response (default true)
```

### SEE ALSO

* [seldon model](seldon_model.md)	 - manage models

