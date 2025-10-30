---
---

# Artifact versions

{% hint style="info" %}
**Note**:  The Seldon CLI allows you to view information about underlying Seldon resources and make changes to them through the scheduler in non-Kubernetes environments. However, it cannot modify underlying manifests within a Kubernetes cluster. Therefore, using the Seldon CLI for control plane operations in a Kubernetes environment is not recommended. For more details, see [Seldon CLI](../getting-started/cli.md).
{% endhint %}

## Seldon V2 Kubernetes Multi Version Artifact Examples

We have a Triton model that has two version folders

Model 1 adds 10 to input, Model 2 multiples by 10 the input. The structure of the artifact repo is shown below:

```sh
config.pbtxt
1/model.py <add 10>
2/model.py <mul 10>

```
## Before you begin

1. Ensure that you have [installed Seldon Core 2](../installation/production-environment/README.md#installing-seldon-core-2) in the namespace `seldon-mesh`.

2. Ensure that you are performing these steps in the directory where you have downloaded the [samples](https://github.com/SeldonIO/seldon-core/tree/v2/samples).

3. Get the IP address of the Seldon Core 2 instance running with Istio:

  ```bash
  ISTIO_INGRESS=$(kubectl get svc seldon-mesh -n seldon-mesh -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

  echo "Seldon Core 2: http://$ISTIO_INGRESS"
  ```
  {% hint style="info" %}
  Make a note of the IP address that is displayed in the output. Replace <INGRESS_IP> with your service mesh's ingress IP address in the following commands.
  {% endhint %}

### Model

```bash
cat ./models/multi-version-1.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: math
spec:
  storageUri: "gs://seldon-models/scv2/samples/triton_23-03/multi-version"
  artifactVersion: 1
  requirements:
  - triton
  - python

```

```bash
kubectl apply -f ./models/multi-version-1.yaml -n seldon-mesh
```

```
model.mlops.seldon.io/math created

```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n seldon-mesh
```

```
model.mlops.seldon.io/math condition met

```

{% tabs %}

{% tab title="curl" %} 

```bash
curl -k http://<INGRESS_IP>:80/v2/models/math/infer \                           
  -H "Host: seldon-mesh.inference.seldon" \
  -H "Content-Type: application/json" \
  -H "Seldon-Model: math" \
  -d '{
    "model_name": "math",
    "inputs": [
      {
        "name": "INPUT",
        "datatype": "FP32",
        "shape": [4],
        "data": [1, 2, 3, 4]
      }
    ]
  }' | jq
```

```bash
{
  "model_name": "math_1",
  "model_version": "1",
  "outputs": [
    {
      "name": "OUTPUT",
      "datatype": "FP32",
      "shape": [
        4
      ],
      "data": [
        11.0,
        12.0,
        13.0,
        14.0
      ]
    }
  ]
}

```
 {% endtab %}

{% tab title="seldon-cli" %} 

```bash
seldon model infer math --inference-mode grpc --inference-host <INGRESS_IP>:80 \
  '{"model_name":"math","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M .
```

```json
{
  "modelName": "math_1",
  "modelVersion": "1",
  "outputs": [
    {
      "name": "OUTPUT",
      "datatype": "FP32",
      "shape": [
        "4"
      ],
      "contents": {
        "fp32Contents": [
          11,
          12,
          13,
          14
        ]
      }
    }
  ]
}

```
 {% endtab %}


{% endtabs %}


```bash
cat ./models/multi-version-2.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: math
spec:
  storageUri: "gs://seldon-models/scv2/samples/triton_23-03/multi-version"
  artifactVersion: 2
  requirements:
  - triton
  - python

```

```bash
kubectl apply -f ./models/multi-version-2.yaml -n seldon-mesh
```

```
model.mlops.seldon.io/math configured

```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n seldon-mesh
```

```
model.mlops.seldon.io/math condition met

```
{% tabs %}

{% tab title="curl" %}

```bash
curl -k http://<INGRESS_IP>:80/v2/models/math/infer \
  -H "Host: seldon-mesh.inference.seldon" \
  -H "Content-Type: application/json" \
  -H "Seldon-Model: math" \
  -d '{
    "model_name": "math",
    "inputs": [
      {
        "name": "INPUT",
        "datatype": "FP32",
        "shape": [4],
        "data": [1, 2, 3, 4]
      }
    ]
  }' | jq
```

```bash
{
  "model_name": "math_2",
  "model_version": "1",
  "outputs": [
    {
      "name": "OUTPUT",
      "datatype": "FP32",
      "shape": [
        4
      ],
      "data": [
        10.0,
        20.0,
        30.0,
        40.0
      ]
    }
  ]
}

```
 {% endtab %}

{% tab title="seldon-cli" %} 

```bash
seldon model infer math --inference-mode grpc --inference-host <INGRESS_IP>:80 \
  '{"model_name":"math","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M .
```

```json
{
  "modelName": "math_2",
  "modelVersion": "1",
  "outputs": [
    {
      "name": "OUTPUT",
      "datatype": "FP32",
      "shape": [
        "4"
      ],
      "contents": {
        "fp32Contents": [
          10,
          20,
          30,
          40
        ]
      }
    }
  ]
}

```
 {% endtab %}


{% endtabs %}


```bash
kubectl delete -f ./models/multi-version-1.yaml -n seldon-mesh
```

```
model.mlops.seldon.io "math" deleted

```
