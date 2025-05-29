---
description: Monitor and check the deployment status of your Seldon Core inference servers using the seldon server status CLI command. This command provides real-time status updates, server health information, and detailed deployment state for your machine learning inference infrastructure.
---

## seldon server status

get status for server

### Synopsis

get the status for a server

```
seldon server status [flags]
```

### Options

```
      --authority string        authority (HTTP/2) or virtual host (HTTP/1)
  -h, --help                    help for status
      --scheduler-host string   seldon scheduler host (default "0.0.0.0:9004")
  -r, --show-request            show request
  -o, --show-response           show response (default true)
```

### SEE ALSO

* [seldon server](seldon_server.md)	 - manage servers

