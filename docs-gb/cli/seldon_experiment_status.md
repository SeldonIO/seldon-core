---
description: Monitor and check the status of your Seldon Core experiments using the seldon experiment status CLI command. This command provides real-time status updates, allows waiting for experiment activation, and displays detailed experiment state information in your Seldon Core deployment.
---

## seldon experiment status

get status for experiment

### Synopsis

get status for experiment

```
seldon experiment status <experimentName> [flags]
```

### Options

```
      --authority string        authority (HTTP/2) or virtual host (HTTP/1)
  -h, --help                    help for status
      --scheduler-host string   seldon scheduler host (default "0.0.0.0:9004")
  -r, --show-request            show request
  -o, --show-response           show response (default true)
  -w, --wait                    wait for experiment to be active
```

### SEE ALSO

* [seldon experiment](seldon_experiment.md)	 - manage experiments

