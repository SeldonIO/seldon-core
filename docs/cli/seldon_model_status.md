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

