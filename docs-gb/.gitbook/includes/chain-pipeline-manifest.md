```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: tfsimples
spec:
  steps:
    - name: tfsimple1
    - name: tfsimple2
      inputs:
      - tfsimple1
      tensorMap:
        tfsimple1.outputs.OUTPUT0: INPUT0
        tfsimple1.outputs.OUTPUT1: INPUT1
  output:
    steps:
    - tfsimple2
```