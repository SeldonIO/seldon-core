---
description: Remove and undeploy machine learning models from Seldon Core using the seldon model unload CLI command. This command helps you safely unload models, clean up model resources, and manage model lifecycle in your Seldon Core deployment.
---

## seldon model unload

unload a model

### Synopsis

unload a model

```
seldon model unload <modelName> [flags]
```

### Options

```
      --authority string        authority (HTTP/2) or virtual host (HTTP/1)
  -f, --file-path string        model manifest file (YAML)
      --force                   force control plane mode (load model, etc.)
  -h, --help                    help for unload
      --scheduler-host string   seldon scheduler host (default "0.0.0.0:9004")
  -v, --verbose                 verbose output
```

### SEE ALSO

* [seldon model](seldon_model.md)	 - manage models

