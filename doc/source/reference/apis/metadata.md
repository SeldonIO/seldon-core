# Model and Deployment Metadata

![metadata](./metadata.svg)



## Incubating feature note
The model metadata feature has currently "incubating" status.
This means that we are currently exploring the best possible interface and functionality for this feature.

As a warning word this means that the API or the way you define metadata may be subject to change before
this feature graduates. If you have any comments or suggestion please open the issue on our GitHub project.

Incubating update 1:
- we added `v1` format that better describes current SeldonMessage
- definition through environmental variables now accepts both `yaml` and `json` input

Incubating update 2:
- we added GRPC support for both Model and Graph metadata
- adjustments to `v1` (it is now also an array of inputs/outputs metadata) and removed explicit distinction

We plan to graduate metadata features with the 1.3 release of Seldon Core.

## Examples
- [Model Metadata (examples)](../../examples/metadata.html)
- [Model Metadata (format examples)](../../examples/metadata_schema.html)
- [Deployment Level Metadata](../../examples/graph-metadata.html)
- [Metadata with GRPC](../../examples/metadata_grpc.html)

- [SKLearn Server example with MinIO](../../examples/minio-sklearn.html)
- [Deploying models trained with Pachyderm](../../examples/pachyderm.html)
- [Deploying models trained with DVC](../../examples/dvc.html)


## Model Metadata (incubating)

With Seldon you can easily add metadata to your models.

### Prepackaged model servers

To add metadata to your prepackaged model servers simply add a file named `metadata.yaml`
to the S3 bucket with your model:
```YAML
name: my-model
versions: [my-model/v1]
platform: platform-name
inputs:
- messagetype: tensor
  schema:
    names: [a, b, c, d]
    shape: [4]
outputs:
- messagetype: tensor
  schema:
    shape: [ 1 ]
```

See [SKLearn Server example with MinIO](../../examples/minio-sklearn.html) for more details.


### Python Language Wrapper

You can add model metadata you your custom Python model by implementing `init_metadata` method:

```python
class Model:
    ...
    def init_metadata(self):
        meta = {
            "name": "my-model-name",
            "versions": ["my-model-version-01"],
            "platform": "seldon",
            "inputs": [
                {
                    "messagetype": "tensor",
                    "schema": {"names": ["a", "b", "c", "d"], "shape": [4]},
                }
            ],
            "outputs": [{"messagetype": "tensor", "schema": {"shape": [1]}}],
        }
        return meta
```

See [Python wrapper](../../python/python_component.html#incubating-features) documentation for more details and
notebook [Basic Examples for Model with Metadata](../../examples/metadata.html).


### Overwrite via environmental variable

You can also always specify `MODEL_METADATA` environmental variable which takes ultimate priority.

```YAML
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: seldon-model
spec:
  name: test-deployment
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - name: my-model
          image: ...
          env:
          - name: MODEL_METADATA
            value: |
              ---
              name: my-model-name
              versions: [ my-model-version ]
              platform: seldon
              inputs:
              - messagetype: tensor
                schema:
                  names: [a, b, c, d]
                  shape: [4]
              outputs:
              - messagetype: tensor
                schema:
                  shape: [ 1 ]
    graph:
      name: my-model
      ...
    name: example
    replicas: 1
```

## Deployment Metadata (incubating)
Model metadata allow you to specify metadata for each of the components (nodes) in your graph.
New orchestrator engine will probe all nodes for their metadata and derive global `inputs` and `outputs` of your graph.
It will then expose them together with all nodes' metadata at a single endpoint `/api/v1.0/metadata/` of your deployment.

![graph-metadata](./graph-metadata.svg)

Example response:
```json
{
    "name": "example",
    "models": {
        "node-one": {
            "name": "node-one",
            "platform": "seldon",
            "versions": ["generic-node/v0.3"],
            "inputs": [
                {"messagetype": "tensor", "schema": {"names": ["one-input"]}}
            ],
            "outputs": [
                {"messagetype": "tensor", "schema": {"names": ["one-output"]}}
            ],
        },
        "node-two": {
            "name": "node-two",
            "platform": "seldon",
            "versions": ["generic-node/v0.3"],
            "inputs": [
                {"messagetype": "tensor", "schema": {"names": ["two-input"]}}
            ],
            "outputs": [
                {"messagetype": "tensor", "schema": {"names": ["two-output"]}}
            ],
        }
    },
    "graphinputs": [
        {"messagetype": "tensor", "schema": {"names": ["one-input"]}}
    ],
    "graphoutputs": [
        {"messagetype": "tensor", "schema": {"names": ["two-output"]}}
    ]
}
```

See example [notebook](../../examples/graph-metadata.html) for more details.



## Metadata endpoint

Model metadata can be obtained through GET request at `/api/v1.0/metadata/{MODEL_NAME}` endpoint of your deployment.

Example response:
```json
{
  "name": "my-model",
  "versions": ["my-model/v1"],
  "platform": "platform-name",
  "inputs": [{"messagetype": "tensor", "schema": {"shape": [1, 5]}}],
  "outputs": [{"messagetype": "tensor", "schema": {"shape": [1, 3]}}]
}
```


## SeldonMessage metadata vs kfserving TensorMetadata

You can define inputs/outputs of your model metadata using one of two formats:
- `v1` format that closely correlates to the current structure of `SeldonMessage`
- `v2` format that is future-proof and fully compatible with [kfserving dataplane proposal](https://github.com/kubeflow/kfserving/blob/master/docs/predict-api/v2/required_api.md#model-metadata) dataplane proposal

### SeldonMessage metadata

#### ndarray input/output
```YAML
name: my-model-name
versions: [ my-model-version-01 ]
platform: seldon
inputs:
- messagetype: ndarray
  schema:
    names: [a, b]
    shape: [ 2, 2 ]
outputs:
- messagetype: ndarray
  schema:
    shape: [ 1 ]
```

This metadata would mean that following two input is valid for this model:
```JSON
{"data": {"names": ["a", "b"], "ndarray": [[1, 2], [3, 4]]}}
```

Note: similar format is valid for messagetype of `tensor` and `tftensor`.

#### jsonData input/output
```YAML
name: my-model-name
versions: [ my-model-version-01 ]
platform: seldon
inputs:
- messagetype: jsonData
  schema:
      type: object
      properties:
          my-names:
              type: array
              items:
                  type: string
          my-data:
            type: array
            items:
                type: number
                format: double
outputs:
- messagetype: ndarray
  schema:
    shape: [ 1 ]
```

Example model input:
```JSON
{"jsonData": {"my-names": ["a", "b", "c"], "my-data": [1.0, 4.2, 3.14]}}
```

The `schema` field is optional and can leaves user total freedom over its structure.

Note: as you can see you can mix inputs and outputs of different types!

#### strData input/output
```YAML
name: my-model-name
versions: [ my-model-version-01 ]
platform: seldon
inputs:
- messagetype: strData
outputs:
- messagetype: strData
```

Example model input:
```JSON
{"strData": "some test input"}
```

#### custom input/output format

You can also specify your custom `messagetype`. In this case there are no restrictions
on keys that you define under the `schema` field. This may be useful for `raw` methods.

```YAML
name: my-model-name
versions: [ my-model-version-01 ]
platform: seldon
inputs:
- messagetype: customData
  schema:
    my-names: ["a", "b", "c"]
outputs:
- messagetype: tensor
  schema:
    shape: [ 1 ]
```


### kfserving TensorMetadata
You can easily define metadata for your models that is compatible with [kfserving dataplane proposal](https://github.com/kubeflow/kfserving/blob/master/docs/predict-api/v2/required_api.md#model-metadata) specification.
```
$metadata_model_response =
{
  "name" : $string,
  "versions" : [ $string, ... ] #optional,
  "platform" : $string,
  "inputs" : [ $metadata_tensor, ... ],
  "outputs" : [ $metadata_tensor, ... ]
}
```
with
```
$metadata_tensor =
{
  "name" : $string,
  "datatype" : $string,
  "shape" : [ $number, ... ]
}
```

Example definition
```YAML
name: my-model-name
versions: [ my-model-version-01 ]
platform: seldon
inputs:
- datatype: BYTES
  name: input
  shape: [ 1, 4 ]
outputs:
- datatype: BYTES
  name: output
  shape: [ 3 ]
```
