---
description: Monitor and check the deployment status of your Seldon Core models using the seldon model status CLI command. This command provides real-time status updates, allows waiting for specific model conditions, and displays detailed deployment state information for your machine learning models.
---

## seldon model status

get status for model

### Synopsis

get the status for a model

```
seldon model status <modelName> [flags]
```

### Options

```
      --authority string        authority (HTTP/2) or virtual host (HTTP/1)
  -h, --help                    help for status
      --scheduler-host string   seldon scheduler host (default "0.0.0.0:9004")
  -r, --show-request            show request
  -o, --show-response           show response (default true)
  -w, --wait string             model wait condition
```

### SEE ALSO

* [seldon model](seldon_model.md)	 - manage models

