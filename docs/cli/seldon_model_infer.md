## seldon model infer

run inference on a model

### Synopsis

call a model with a given input and get a prediction

```
seldon model infer <modelName> (data) [flags]
```

### Options

```
      --authority string        authority (HTTP/2) or virtual host (HTTP/1)
  -f, --file-path string        inference payload file
      --header stringArray      add a header, e.g. key=value; use the flag multiple times to add more than one header
  -h, --help                    help for infer
      --inference-host string   seldon inference host (default "0.0.0.0:9000")
      --inference-mode string   inference mode (rest or grpc) (default "rest")
  -i, --iterations int          how many times to run inference (default 1)
  -t, --seconds int             number of secs to run inference
      --show-headers            show request and response headers
  -r, --show-request            show request
  -o, --show-response           show response (default true)
  -s, --sticky-session          use sticky session from last inference (only works with experiments)
```

### SEE ALSO

* [seldon model](seldon_model.md)	 - manage models

