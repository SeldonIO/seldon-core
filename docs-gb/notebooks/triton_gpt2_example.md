# Pretrained  GPT2  Model Deployment Example

In this notebook, we will run an example of text generation using GPT2 model exported from HuggingFace and deployed with Seldon's Triton pre-packed server. the example also covers converting the model to ONNX format.
The implemented example below is of the Greedy approach for the next token prediction.
more info: https://huggingface.co/transformers/model_doc/gpt2.html?highlight=gpt2

After we have the module deployed to Kubernetes, we will run a simple load test to evaluate the module inference performance.


## Steps:
1. Download pretrained GPT2 model from hugging face
2. Convert the model to ONNX
3. Store it in MinIo bucket
4. Setup Seldon-Core in your kubernetes cluster
5. Deploy the ONNX model with Seldon’s prepackaged Triton server.
6. Interact with the model, run a greedy alg example (generate sentence completion)
7. Run load test using vegeta
8. Clean-up

## Basic requirements
* Helm v3.0.0+
* A Kubernetes cluster running v1.13 or above (minkube / docker-for-windows work well if enough RAM)
* kubectl v1.14+
* Python 3.6+ 


```python
%%writefile requirements.txt
transformers==4.5.1
torch==1.8.1
tokenizers<0.11,>=0.10.1
tensorflow==2.4.1
tf2onnx
```

    Writing requirements.txt



```python
!pip install --trusted-host=pypi.python.org --trusted-host=pypi.org --trusted-host=files.pythonhosted.org -r requirements.txt
```

### Export HuggingFace TFGPT2LMHeadModel pre-trained model and save it locally


```python
from transformers import GPT2Tokenizer, TFGPT2LMHeadModel

tokenizer = GPT2Tokenizer.from_pretrained("gpt2")
model = TFGPT2LMHeadModel.from_pretrained(
    "gpt2", from_pt=True, pad_token_id=tokenizer.eos_token_id
)
model.save_pretrained("./tfgpt2model", saved_model=True)
```

### Convert the TensorFlow saved model to ONNX


```python
!python -m tf2onnx.convert --saved-model ./tfgpt2model/saved_model/1 --opset 11  --output model.onnx
```

### Copy your model to a local MinIo
#### Setup MinIo
Use the provided [notebook](../notebooks/minio_setup.md) to install MinIo in your cluster and configure `mc` CLI tool. Instructions also [online](https://docs.min.io/docs/minio-client-quickstart-guide.html).

-- Note: You can use your prefer remote storage server (google/ AWS etc.)

#### Create a Bucket and store your model


```python
!mc mb minio/language-models/onnx-gpt2/1 -p
!mc cp ./model.onnx minio/language-models/onnx-gpt2/1/
```

    [m[32;1mBucket created successfully `minio/language-models/onnx-gpt2/1`.[0m
    ...odel.onnx:  622.29 MiB / 622.29 MiB ┃▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓┃ 136.59 MiB/s 4s[0m[0m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m[m[32;1m

### Run Seldon in your kubernetes cluster

Follow the [Seldon-Core Setup notebook](../notebooks/seldon-core-setup.md) to Setup a cluster with Ambassador Ingress or Istio and install Seldon Core

### Deploy your model with Seldon pre-packaged Triton server


```python
%%writefile secret.yaml

apiVersion: v1
kind: Secret
metadata:
  name: seldon-init-container-secret
type: Opaque
stringData:
  RCLONE_CONFIG_S3_TYPE: s3
  RCLONE_CONFIG_S3_PROVIDER: minio
  RCLONE_CONFIG_S3_ENV_AUTH: "false"
  RCLONE_CONFIG_S3_ACCESS_KEY_ID: minioadmin
  RCLONE_CONFIG_S3_SECRET_ACCESS_KEY: minioadmin
  RCLONE_CONFIG_S3_ENDPOINT: http://minio.minio-system.svc.cluster.local:9000

```

    Writing secret.yaml



```python
%%writefile gpt2-deploy.yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: gpt2
spec:
  predictors:
  - graph:
      implementation: TRITON_SERVER
      logger:
        mode: all
      modelUri: s3://language-models
      envSecretRefName: seldon-init-container-secret
      name: onnx-gpt2
      type: MODEL
    name: default
    replicas: 1
  protocol: kfserving
```

    Writing gpt2-deploy.yaml



```python
!kubectl apply -f secret.yaml -n default
!kubectl apply -f gpt2-deploy.yaml -n default
```

    secret/seldon-init-container-secret configured
    seldondeployment.machinelearning.seldon.io/gpt2 unchanged



```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=gpt2 -o jsonpath='{.items[0].metadata.name}')
```

    deployment "gpt2-default-0-onnx-gpt2" successfully rolled out


#### Interact with the model: get model metadata (a "test" request to make sure our model is available and loaded correctly)


```python
!curl -s http://localhost:80/seldon/default/gpt2/v2/models/onnx-gpt2
```

    {"name":"onnx-gpt2","versions":["1"],"platform":"onnxruntime_onnx","inputs":[{"name":"input_ids","datatype":"INT32","shape":[-1,-1]},{"name":"attention_mask","datatype":"INT32","shape":[-1,-1]}],"outputs":[{"name":"past_key_values","datatype":"FP32","shape":[12,2,-1,12,-1,64]},{"name":"logits","datatype":"FP32","shape":[-1,-1,50257]}]}

### Run prediction test: generate a sentence completion using GPT2 model  - Greedy approach



```python
import json

import numpy as np
import requests
from transformers import GPT2Tokenizer

tokenizer = GPT2Tokenizer.from_pretrained("gpt2")
input_text = "I enjoy working in Seldon"
count = 0
max_gen_len = 10
gen_sentence = input_text
while count < max_gen_len:
    input_ids = tokenizer.encode(gen_sentence, return_tensors="tf")
    shape = input_ids.shape.as_list()
    payload = {
        "inputs": [
            {
                "name": "input_ids",
                "datatype": "INT32",
                "shape": shape,
                "data": input_ids.numpy().tolist(),
            },
            {
                "name": "attention_mask",
                "datatype": "INT32",
                "shape": shape,
                "data": np.ones(shape, dtype=np.int32).tolist(),
            },
        ]
    }

    ret = requests.post(
        "http://localhost:80/seldon/default/gpt2/v2/models/onnx-gpt2/infer",
        json=payload,
    )

    try:
        res = ret.json()
    except:
        continue

    # extract logits
    logits = np.array(res["outputs"][1]["data"])
    logits = logits.reshape(res["outputs"][1]["shape"])

    # take the best next token probability of the last token of input ( greedy approach)
    next_token = logits.argmax(axis=2)[0]
    next_token_str = tokenizer.decode(
        next_token[-1:], skip_special_tokens=True, clean_up_tokenization_spaces=True
    ).strip()
    gen_sentence += " " + next_token_str
    count += 1

print(f"Input: {input_text}\nOutput: {gen_sentence}")
```

    Input: I enjoy working in Seldon
    Output: I enjoy working in Seldon 's office , and I 'm glad to see that


### Run Load Test / Performance Test using vegeta

#### Install vegeta, for more details take a look in [vegeta](https://github.com/tsenart/vegeta#install) official documentation


```python
!wget https://github.com/tsenart/vegeta/releases/download/v12.8.3/vegeta-12.8.3-linux-amd64.tar.gz
!tar -zxvf vegeta-12.8.3-linux-amd64.tar.gz
!chmod +x vegeta
```

#### Generate vegeta [target file](https://github.com/tsenart/vegeta#-targets) contains "post" cmd with payload in the requiered structure


```python
import base64
import json
from subprocess import PIPE, Popen, run

import numpy as np
from transformers import GPT2Tokenizer, TFGPT2LMHeadModel

tokenizer = GPT2Tokenizer.from_pretrained("gpt2")
input_text = "I enjoy working in Seldon"
input_ids = tokenizer.encode(input_text, return_tensors="tf")
shape = input_ids.shape.as_list()
payload = {
    "inputs": [
        {
            "name": "input_ids",
            "datatype": "INT32",
            "shape": shape,
            "data": input_ids.numpy().tolist(),
        },
        {
            "name": "attention_mask",
            "datatype": "INT32",
            "shape": shape,
            "data": np.ones(shape, dtype=np.int32).tolist(),
        },
    ]
}

cmd = {
    "method": "POST",
    "header": {"Content-Type": ["application/json"]},
    "url": "http://localhost:80/seldon/default/gpt2/v2/models/gpt2/infer",
    "body": base64.b64encode(bytes(json.dumps(payload), "utf-8")).decode("utf-8"),
}

with open("vegeta_target.json", mode="w") as file:
    json.dump(cmd, file)
    file.write("\n\n")
```


```python
!vegeta attack -targets=vegeta_target.json -rate=1 -duration=60s -format=json | vegeta report -type=text
```

### Clean-up


```python
!kubectl delete -f gpt2-deploy.yaml -n default
```

    seldondeployment.machinelearning.seldon.io "gpt2" deleted

