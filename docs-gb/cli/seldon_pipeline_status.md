---
description: Monitor and check the deployment status of your Seldon Core pipelines using the seldon pipeline status CLI command. This command provides real-time status updates, allows waiting for specific pipeline conditions, and displays detailed deployment state information for your data processing pipelines.
---

## seldon pipeline status

status of a pipeline

### Synopsis

status of a pipeline

```
seldon pipeline status <pipelineName> [flags]
```

### Options

```
      --authority string        authority (HTTP/2) or virtual host (HTTP/1)
  -h, --help                    help for status
      --scheduler-host string   seldon scheduler host (default "0.0.0.0:9004")
  -r, --show-request            show request
  -o, --show-response           show response (default true)
  -w, --wait string             pipeline wait condition
```

### SEE ALSO

* [seldon pipeline](seldon_pipeline.md)	 - manage pipelines

