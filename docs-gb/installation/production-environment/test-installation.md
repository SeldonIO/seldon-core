# Verify the Installation of Seldon Core 2

To confirm the successful installation of [Seldon Core 2](/docs-gb/installation/production-environment/README.md), [Kafka](/docs-gb/installation/production-environment/kafka/), and the [service mesh](/docs-gb/installation/production-environment/ingress-controller/), deploy a sample model and perform an inference test. Follow these steps:

## Deploy the Iris Model

1. Apply the following configuration to deploy the Iris model in the namespace `seldon-mesh`:

```bash
kubectl apply -f - --namespace=seldon-mesh <<EOF
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: iris
spec:
  storageUri: "gs://seldon-models/scv2/samples/mlserver_1.3.5/iris-sklearn"
  requirements:
    - sklearn
EOF

```
The output is:
```
model.mlops.seldon.io/iris configured
```
2. Verify that the model is deployed in the namespace `seldon-mesh`.
 ```bash
 kubectl wait --for condition=ready --timeout=300s model --all -n seldon-mesh
 ```
 When the model is deployed, the output is similar to:
 ```bash
 model.mlops.seldon.io/iris condition met
 ``` 
## Perform an Inference test

1. Use curl to send a test inference request to the deployed model. Replace <INGRESS_IP> with your service mesh's ingress IP address.
Ensure that:
* The Host header matches the expected virtual host configured in your service mesh.
* The Seldon-Model header specifies the correct model name.

```bash
curl -k http://<INGRESS_IP>:80/v2/models/iris/infer \
  -H "Host: seldon-mesh.inference.seldon" \
  -H "Content-Type: application/json" \
  -H "Seldon-Model: iris" \
  -d '{
    "inputs": [
      {
        "name": "predict",
        "shape": [1, 4],
        "datatype": "FP32",
        "data": [[1, 2, 3, 4]]
      }
    ]
  }'
```

The output is similar to:
```bash
{"model_name":"iris_1","model_version":"1","id":"f4d8b82f-2af3-44fb-b115-60a269cbfa5e","parameters":{},"outputs":[{"name":"predict","shape":[1,1],"datatype":"INT64","parameters":{"content_type":"np"},"data":[2]}]}
```
  


