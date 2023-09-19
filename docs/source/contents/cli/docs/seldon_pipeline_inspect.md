## seldon pipeline inspect

inspect data in a pipeline

### Synopsis

inspect data in a pipeline. Specify as pipelineName or pipelineName.(inputs|outputs) or pipelineName.stepName or pipelineName.stepName.(inputs|outputs) or pipelineName.stepName.(inputs|outputs).tensorName

```
seldon pipeline inspect <expression> [flags]
```

### Options

```
      --format string           inspect output format: raw or json. Default raw (default "raw")
  -h, --help                    help for inspect
      --kafka-broker string     kafka broker (default "0.0.0.0:9092")
      --namespace string        Kubernetes namespace. Default default (default "default")
      --offset int              message offset to start reading from, i.e. default 1 is the last message only (default 1)
      --request-id string       request id to show, if not specified will be all messages in offset range
      --scheduler-host string   seldon scheduler host (default "0.0.0.0:9004")
      --verbose                 display more details, such as headers
```

### SEE ALSO

* [seldon pipeline](seldon_pipeline.md)	 - manage pipelines

