## Huggingface Speech to Sentiment Pipeline Example

In this example we create a Pipeline to chain two huggingface models to allow speech to sentiment functionalityand add an explainer to understand the result.

This example also illustrates how explainers can target pipelines to allow complex explanations flows.

![architetcure](speech-to-sentiment.jpg)

This example requires **ffmpeg** package to be installed locally. run `make install-requirements` for the Python dependencies.




```python
from ipywebrtc import AudioRecorder, CameraStream
import torchaudio
from IPython.display import Audio
import base64
import json
import requests
import os
import time
```

Create a method to load speech from recorder; transform into mp3 and send at base64 data. On return of the result extract and show the text and sentiment.


```python
reqJson = json.loads('{"inputs":[{"name":"args", "parameters": {"content_type": "base64"}, "data":[],"datatype":"BYTES","shape":[1]}]}')
url = "http://0.0.0.0:9000/v2/models/model/infer"
def infer(resource: str):
    with open('recording.webm', 'wb') as f:
        f.write(recorder.audio.value)
    !ffmpeg -i recording.webm -vn -ab 128k -ar 44100 file.mp3 -y -hide_banner -loglevel panic
    with open("file.mp3", mode='rb') as file:
        fileContent = file.read()
        encoded = base64.b64encode(fileContent)
        base64_message = encoded.decode('utf-8')
    reqJson["inputs"][0]["data"] = [str(base64_message)]
    headers = {"Content-Type": "application/json", "seldon-model": resource}
    response_raw = requests.post(url, json=reqJson, headers=headers)
    j = response_raw.json()
    sentiment = j["outputs"][0]["data"][0]
    text = j["outputs"][1]["data"][0]
    reqId = response_raw.headers["x-request-id"]
    print(reqId)
    os.environ["REQUEST_ID"]=reqId
    print(text)
    print(sentiment)
```

### Load Huggingface Models

We will load two Huggingface models for speech to text and text to sentiment.


```python
!cat ../../models/hf-whisper.yaml
!echo "---"
!cat ../../models/hf-sentiment.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: whisper
    spec:
      storageUri: "gs://seldon-models/mlserver/huggingface/whisper"
      requirements:
      - huggingface
    ---
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: sentiment
    spec:
      storageUri: "gs://seldon-models/mlserver/huggingface/sentiment"
      requirements:
      - huggingface



```python
!seldon model load -f ../../models/hf-whisper.yaml
!seldon model load -f ../../models/hf-sentiment.yaml
```

    {}
    {}



```python
!seldon model status whisper -w ModelAvailable | jq -M .
!seldon model status sentiment -w ModelAvailable | jq -M .
```

    {}
    {}


### Create Explain Pipeline

To allow Alibi-Explain to more easily explain the sentiment we will need:

 * input and output transfrorms that take the Dict values input and output by the Huggingface sentiment model and turn them into values that Alibi-Explain can easily understand with the core values we want to explain and the outputs from the sentiment model.
 * A separate Pipeline to allow us to join the sentiment model with the output transform

These transform models are MLServer custom runtimes as shown below:


```python
!cat ./sentiment-input-transform/model.py | pygmentize
```

    [37m# Copyright 2022 Seldon Technologies Ltd.[39;49;00m
    [37m#[39;49;00m
    [37m# Licensed under the Apache License, Version 2.0 (the "License");[39;49;00m
    [37m# you may not use this file except in compliance with the License.[39;49;00m
    [37m# You may obtain a copy of the License at[39;49;00m
    [37m#[39;49;00m
    [37m#    http://www.apache.org/licenses/LICENSE-2.0[39;49;00m
    [37m#[39;49;00m
    [37m# Unless required by applicable law or agreed to in writing, software[39;49;00m
    [37m# distributed under the License is distributed on an "AS IS" BASIS,[39;49;00m
    [37m# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.[39;49;00m
    [37m# See the License for the specific language governing permissions and[39;49;00m
    [37m# limitations under the License.[39;49;00m
    
    [34mfrom[39;49;00m [04m[36mmlserver[39;49;00m [34mimport[39;49;00m MLModel
    [34mfrom[39;49;00m [04m[36mmlserver[39;49;00m[04m[36m.[39;49;00m[04m[36mtypes[39;49;00m [34mimport[39;49;00m InferenceRequest, InferenceResponse, ResponseOutput
    [34mfrom[39;49;00m [04m[36mmlserver[39;49;00m[04m[36m.[39;49;00m[04m[36mcodecs[39;49;00m[04m[36m.[39;49;00m[04m[36mstring[39;49;00m [34mimport[39;49;00m StringRequestCodec
    [34mfrom[39;49;00m [04m[36mmlserver[39;49;00m[04m[36m.[39;49;00m[04m[36mlogging[39;49;00m [34mimport[39;49;00m logger
    [34mimport[39;49;00m [04m[36mjson[39;49;00m
    
    
    [34mclass[39;49;00m [04m[32mSentimentInputTransformRuntime[39;49;00m(MLModel):
    
      [34masync[39;49;00m [34mdef[39;49;00m [32mload[39;49;00m([36mself[39;49;00m) -> [36mbool[39;49;00m:
        [34mreturn[39;49;00m [36mself[39;49;00m.ready
    
      [34masync[39;49;00m [34mdef[39;49;00m [32mpredict[39;49;00m([36mself[39;49;00m, payload: InferenceRequest) -> InferenceResponse:
        res_list = [36mself[39;49;00m.decode_request(payload, default_codec=StringRequestCodec)
        texts = []
        [34mfor[39;49;00m res [35min[39;49;00m res_list:
          logger.debug([33m"[39;49;00m[33mdecoded data: [39;49;00m[33m%s[39;49;00m[33m"[39;49;00m, res)
          text = json.loads(res)
          texts.append(text[[33m"[39;49;00m[33mtext[39;49;00m[33m"[39;49;00m])
    
        [34mreturn[39;49;00m StringRequestCodec.encode_response(
          model_name=[33m"[39;49;00m[33msentiment[39;49;00m[33m"[39;49;00m,
          payload=texts
        )



```python
!cat ./sentiment-output-transform/model.py | pygmentize
```

    [37m# Copyright 2022 Seldon Technologies Ltd.[39;49;00m
    [37m#[39;49;00m
    [37m# Licensed under the Apache License, Version 2.0 (the "License");[39;49;00m
    [37m# you may not use this file except in compliance with the License.[39;49;00m
    [37m# You may obtain a copy of the License at[39;49;00m
    [37m#[39;49;00m
    [37m#    http://www.apache.org/licenses/LICENSE-2.0[39;49;00m
    [37m#[39;49;00m
    [37m# Unless required by applicable law or agreed to in writing, software[39;49;00m
    [37m# distributed under the License is distributed on an "AS IS" BASIS,[39;49;00m
    [37m# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.[39;49;00m
    [37m# See the License for the specific language governing permissions and[39;49;00m
    [37m# limitations under the License.[39;49;00m
    
    [34mfrom[39;49;00m [04m[36mmlserver[39;49;00m [34mimport[39;49;00m MLModel
    [34mfrom[39;49;00m [04m[36mmlserver[39;49;00m[04m[36m.[39;49;00m[04m[36mtypes[39;49;00m [34mimport[39;49;00m InferenceRequest, InferenceResponse, ResponseOutput
    [34mfrom[39;49;00m [04m[36mmlserver[39;49;00m[04m[36m.[39;49;00m[04m[36mcodecs[39;49;00m [34mimport[39;49;00m StringCodec, Base64Codec, NumpyRequestCodec
    [34mfrom[39;49;00m [04m[36mmlserver[39;49;00m[04m[36m.[39;49;00m[04m[36mcodecs[39;49;00m[04m[36m.[39;49;00m[04m[36mstring[39;49;00m [34mimport[39;49;00m StringRequestCodec
    [34mfrom[39;49;00m [04m[36mmlserver[39;49;00m[04m[36m.[39;49;00m[04m[36mcodecs[39;49;00m[04m[36m.[39;49;00m[04m[36mnumpy[39;49;00m [34mimport[39;49;00m NumpyRequestCodec
    [34mimport[39;49;00m [04m[36mbase64[39;49;00m
    [34mfrom[39;49;00m [04m[36mmlserver[39;49;00m[04m[36m.[39;49;00m[04m[36mlogging[39;49;00m [34mimport[39;49;00m logger
    [34mimport[39;49;00m [04m[36mnumpy[39;49;00m [34mas[39;49;00m [04m[36mnp[39;49;00m
    [34mimport[39;49;00m [04m[36mjson[39;49;00m
    
    [34mclass[39;49;00m [04m[32mSentimentOutputTransformRuntime[39;49;00m(MLModel):
    
      [34masync[39;49;00m [34mdef[39;49;00m [32mload[39;49;00m([36mself[39;49;00m) -> [36mbool[39;49;00m:
        [34mreturn[39;49;00m [36mself[39;49;00m.ready
    
      [34masync[39;49;00m [34mdef[39;49;00m [32mpredict[39;49;00m([36mself[39;49;00m, payload: InferenceRequest) -> InferenceResponse:
        res_list = [36mself[39;49;00m.decode_request(payload, default_codec=StringRequestCodec)
        scores = []
        [34mfor[39;49;00m res [35min[39;49;00m res_list:
          logger.debug([33m"[39;49;00m[33mdecoded data: [39;49;00m[33m%s[39;49;00m[33m"[39;49;00m,res)
          sentiment = json.[34mloads[39;49;00m(res)
          [34mif[39;49;00m sentiment[[33m"[39;49;00m[33mlabel[39;49;00m[33m"[39;49;00m] == [33m"[39;49;00m[33mPOSITIVE[39;49;00m[33m"[39;49;00m:
            scores.[34mappend[39;49;00m([34m1[39;49;00m)
          [34melse[39;49;00m:
            scores.[34mappend[39;49;00m([34m0[39;49;00m)
        [34mreturn[39;49;00m NumpyRequestCodec.encode_response(
          model_name=[33m"[39;49;00m[33msentiments[39;49;00m[33m"[39;49;00m,
          payload=np.[34marray[39;49;00m(scores)
        )



```python
!cat ../../models/hf-sentiment-input-transform.yaml
!echo "---"
!cat ../../models/hf-sentiment-output-transform.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: sentiment-input-transform
    spec:
      storageUri: "gs://seldon-models/scv2/examples/huggingface/sentiment-input-transform"
      requirements:
      - mlserver
      - python
    ---
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: sentiment-output-transform
    spec:
      storageUri: "gs://seldon-models/scv2/examples/huggingface/sentiment-output-transform"
      requirements:
      - mlserver
      - python



```python
!seldon model load -f ../../models/hf-sentiment-input-transform.yaml
!seldon model load -f ../../models/hf-sentiment-output-transform.yaml
```

    {}
    {}



```python
!seldon model status sentiment-input-transform -w ModelAvailable | jq -M .
!seldon model status sentiment-output-transform -w ModelAvailable | jq -M .
```

    {}
    {}



```python
!cat ../../pipelines/sentiment-explain.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: sentiment-explain
    spec:
      steps:
        - name: sentiment
          tensorMap:
            sentiment-explain.inputs.predict: args
        - name: sentiment-output-transform
          inputs:
          - sentiment
      output:
        steps:
        - sentiment-output-transform



```python
!seldon pipeline load -f ../../pipelines/sentiment-explain.yaml
```

    {}



```python
!seldon pipeline status sentiment-explain -w PipelineReady| jq -M .
```

    {
      "pipelineName": "sentiment-explain",
      "versions": [
        {
          "pipeline": {
            "name": "sentiment-explain",
            "uid": "cdrn2c5jr36s73eg9om0",
            "version": 1,
            "steps": [
              {
                "name": "sentiment",
                "tensorMap": {
                  "sentiment-explain.inputs.predict": "args"
                }
              },
              {
                "name": "sentiment-output-transform",
                "inputs": [
                  "sentiment.outputs"
                ]
              }
            ],
            "output": {
              "steps": [
                "sentiment-output-transform.outputs"
              ]
            },
            "kubernetesMeta": {}
          },
          "state": {
            "pipelineVersion": 1,
            "status": "PipelineReady",
            "reason": "created pipeline",
            "lastChangeTimestamp": "2022-11-18T11:49:04.697311598Z"
          }
        }
      ]
    }



```python
!cat ../../models/hf-sentiment-explainer.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: sentiment-explainer
    spec:
      storageUri: "gs://seldon-models/scv2/examples/huggingface/speech-sentiment/explainer"
      explainer:
        type: anchor_text
        pipelineRef: sentiment-explain



```python
!seldon model load -f ../../models/hf-sentiment-explainer.yaml
```

    {}



```python
!seldon model status sentiment-explainer -w ModelAvailable | jq -M .
```

    {}


### Speech to Sentiment Pipeline with Explanation

We can now create the final pipeline that will take speech and generate sentiment alongwith an explanation of why that sentiment was predicted.


```python
!cat ../../pipelines/speech-to-sentiment.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: speech-to-sentiment
    spec:
      steps:
        - name: whisper
        - name: sentiment
          inputs:
          - whisper
          tensorMap:
            whisper.outputs.output: args
        - name: sentiment-input-transform
          inputs:
          - whisper
        - name: sentiment-explainer
          inputs:
          - sentiment-input-transform
      output:
        steps:
        - sentiment
        - whisper



```python
!seldon pipeline load -f ../../pipelines/speech-to-sentiment.yaml
```

    {}



```python
!seldon pipeline status speech-to-sentiment -w PipelineReady| jq -M .
```

    {
      "pipelineName": "speech-to-sentiment",
      "versions": [
        {
          "pipeline": {
            "name": "speech-to-sentiment",
            "uid": "cdrn2htjr36s73eg9omg",
            "version": 1,
            "steps": [
              {
                "name": "sentiment",
                "inputs": [
                  "whisper.outputs"
                ],
                "tensorMap": {
                  "whisper.outputs.output": "args"
                }
              },
              {
                "name": "sentiment-explainer",
                "inputs": [
                  "sentiment-input-transform.outputs"
                ]
              },
              {
                "name": "sentiment-input-transform",
                "inputs": [
                  "whisper.outputs"
                ]
              },
              {
                "name": "whisper"
              }
            ],
            "output": {
              "steps": [
                "sentiment.outputs",
                "whisper.outputs"
              ]
            },
            "kubernetesMeta": {}
          },
          "state": {
            "pipelineVersion": 1,
            "status": "PipelineReady",
            "reason": "created pipeline",
            "lastChangeTimestamp": "2022-11-18T11:49:27.793654254Z"
          }
        }
      ]
    }


### Test


```python
camera = CameraStream(constraints={'audio': True,'video':False})
recorder = AudioRecorder(stream=camera)
recorder
```


    AudioRecorder(audio=Audio(value=b'', format='webm'), stream=CameraStream(constraints={'audio': True, 'video': …



```python
infer("speech-to-sentiment.pipeline")
```

    cdrn3iits0ps73bgtaeg
    {"text": " Cambridge is wonderful and beautiful."}
    {"label": "POSITIVE", "score": 0.9998691082000732}


We will wait for the explanation which is run asynchronously to the functional output from the Pipeline above.


```python
while True:
    base64Res = !seldon pipeline inspect speech-to-sentiment.sentiment-explainer.outputs --format json \
          --request-id ${REQUEST_ID}
    j = json.loads(base64Res[0])
    if j["topics"][0]["msgs"] is not None:
        expBase64 = j["topics"][0]["msgs"][0]["value"]["outputs"][0]["contents"]["bytesContents"][0]
        expRaw = base64.b64decode(expBase64)
        exp = json.loads(expRaw)
        print("")
        print("Explanation anchors:",exp["data"]["anchor"])
        break
    else:
        print(".",end='')
        time.sleep(1)
    
```

    .....
    Explanation anchors: ['beautiful']


### Cleanup


```python
!seldon pipeline unload speech-to-sentiment
!seldon pipeline unload sentiment-explain
```

    {}
    {}



```python
!seldon model unload whisper
!seldon model unload sentiment
!seldon model unload sentiment-explainer
!seldon model unload sentiment-transform
```

    {}
    {}
    {}
    {}



```python

```
