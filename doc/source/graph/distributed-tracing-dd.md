# Distributed Tracing with Datadog

You can use Open Tracing to trace your API calls to Seldon Core. By default Jaeger is supported ([see here](distributed-tracing.md)), but Datadog can also be used. 

Datadog is only supported in the Executor and the Python wrapper at this time.

## Install Datadog

You will need to install the Datadog Agent on your Kubernetes cluster. Follow their [documentation](https://docs.datadoghq.com/agent/kubernetes/?tab=helm). Ensure that APM is enabled in the deployment, see [here](https://docs.datadoghq.com/agent/kubernetes/apm/?tab=helm).

## Configuration

You will need to annotate your Seldon Deployment resource with environment variables to make tracing active and set the appropriate Datadog configuration variables.

  * For each Seldon component you run (e.g., model transformer etc.) you will need to add environment variables to the container section.


### Python Wrapper Configuration

Add an environment variable: `TRACING` with value `1` to activate tracing.

To ensure that Datadog tracing is used, set the environment variable `DD_ENABLED` to `1`

For a complete list of available environment variables, see the [Datadogs python documentation](https://docs.datadoghq.com/tracing/setup/python/#configuration) for the model wrapper, and [Datadogs Go documentation](https://docs.datadoghq.com/tracing/setup/go/) for the executor, but the relevant ones are below:
* `DD_AGENT_HOST=<host>` (defaults to `localhost`)
* `DATADOG_TRACE_AGENT_PORT=<port>` (defaults to `8126`)
* `DD_SERVICE=<svc>` (will default to either `executor`, or the name of your Python class)
* `DD_TAGS=<key:value,key2:value2`
* `DD_SAMPLE_RATE:1` (defaults to 1, keeping all traces)
    * _NOTE: This is a non-standard environment variable, meaning its specific to Seldon._



An example is show below:

```yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: dd-tracing-example
  namespace: seldon
spec:
  name: dd-tracing-example
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - env:
          - name: TRACING
            value: '1'          
          - name: DD_ENABLED
            value: '1'
          - name: DD_AGENT_HOST
            valueFrom:
              fieldRef:
                fieldPath: status.hostIP
          - name: DATADOG_TRACE_AGENT_PORT
            value: '8126'
          - name: DD_SAMPLE_RATE
            value: 0.75 # Keep 75% of traces
          image: seldonio/mock_classifier_rest:1.3
          name: model1
        terminationGracePeriodSeconds: 1
    graph:
      children: []
      endpoint:
        type: REST
      name: model1
      type: MODEL
    name: tracing
    replicas: 1
    svcOrchSpec:
      env:
      - name: TRACING
        value: '1'
      - name: DD_ENABLED
        value: '1'
      - name: DD_AGENT_HOST
        valueFrom:
          fieldRef:
            fieldPath: status.hostIP
      - name: DATADOG_TRACE_AGENT_PORT
        value: '8126'
      - name: DD_SAMPLE_RATE
        value: 0.9 # Keep 90% of traces
```

