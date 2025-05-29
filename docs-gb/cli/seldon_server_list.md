---
description: View and enumerate all available Seldon Core inference servers using the seldon server list CLI command. This command provides a comprehensive overview of deployed servers, their replicas, and currently loaded machine learning models in your Seldon Core environment.
---

## seldon server list

get list of servers

### Synopsis

get the available servers, their replicas and loaded models

```
seldon server list [flags]
```

### Options

```
      --authority string        authority (HTTP/2) or virtual host (HTTP/1)
  -h, --help                    help for list
      --scheduler-host string   seldon scheduler host (default "0.0.0.0:9004")
```

### SEE ALSO

* [seldon server](seldon_server.md)	 - manage servers

