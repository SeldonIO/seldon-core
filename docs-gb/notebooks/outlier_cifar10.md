# Cifar10 Outlier Detection
![demo](../images/demo.png)

In this example we will deploy an image classification model along with an outlier detector trained on the same dataset. For in depth details on creating an outlier detection model for your own dataset see the [alibi-detect project](https://github.com/SeldonIO/alibi-detect) and associated [documentation](https://docs.seldon.ai/alibi-detect). You can find details for this [CIFAR10 example in their documentation](https://docs.seldon.ai/alibi-detect/outlier-detection/examples/od_vae_cifar10) as well.


Prequisites:

  * [Knative eventing installed](https://knative.dev/docs/install/)
    * Ensure the istio-ingressgateway is exposed as a loadbalancer (no auth in this demo)
  * [Seldon Core installed](https://docs.seldon.ai/seldon-core-1/getting-started/installation/installation) 
    * Ensure you install for istio, e.g. for the helm chart `--set istio.enabled=true`
    
    Tested on GKE and Kind with Knative 1.10.1 and Istio 1.16.2


```python
!pip install -r requirements_notebook.txt
```

    Requirement already satisfied: alibi-detect>=0.13.0 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from -r requirements_notebook.txt (line 1)) (0.13.0)
    Requirement already satisfied: matplotlib>=3.1.1 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from -r requirements_notebook.txt (line 2)) (3.10.7)
    Requirement already satisfied: numpy<2.0.0,>=1.16.2 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (1.26.4)
    Requirement already satisfied: pandas<3.0.0,>=1.0.0 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (2.3.3)
    Requirement already satisfied: Pillow<11.0.0,>=5.4.1 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (10.4.0)
    Requirement already satisfied: opencv-python<5.0.0,>=3.2.0 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (4.11.0.86)
    Requirement already satisfied: scipy<2.0.0,>=1.5.0 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (1.16.3)
    Requirement already satisfied: scikit-image<0.25,>=0.19 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (0.22.0)
    Requirement already satisfied: scikit-learn<2.0.0,>=0.20.2 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (1.7.2)
    Requirement already satisfied: transformers<5.0.0,>=4.0.0 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (4.57.1)
    Requirement already satisfied: dill<0.4.0,>=0.3.0 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (0.3.2)
    Requirement already satisfied: tqdm<5.0.0,>=4.28.1 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (4.67.1)
    Requirement already satisfied: requests<3.0.0,>=2.21.0 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (2.32.5)
    Requirement already satisfied: pydantic<3.0.0,>=1.8.0 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (2.12.4)
    Requirement already satisfied: toml<1.0.0,>=0.10.1 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (0.10.2)
    Requirement already satisfied: catalogue<3.0.0,>=2.0.0 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (2.0.8)
    Requirement already satisfied: numba!=0.54.0,<0.60.0,>=0.50.0 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (0.59.1)
    Requirement already satisfied: typing-extensions>=3.7.4.3 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (4.15.0)
    Requirement already satisfied: contourpy>=1.0.1 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from matplotlib>=3.1.1->-r requirements_notebook.txt (line 2)) (1.3.3)
    Requirement already satisfied: cycler>=0.10 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from matplotlib>=3.1.1->-r requirements_notebook.txt (line 2)) (0.12.1)
    Requirement already satisfied: fonttools>=4.22.0 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from matplotlib>=3.1.1->-r requirements_notebook.txt (line 2)) (4.60.1)
    Requirement already satisfied: kiwisolver>=1.3.1 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from matplotlib>=3.1.1->-r requirements_notebook.txt (line 2)) (1.4.9)
    Requirement already satisfied: packaging>=20.0 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from matplotlib>=3.1.1->-r requirements_notebook.txt (line 2)) (25.0)
    Requirement already satisfied: pyparsing>=3 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from matplotlib>=3.1.1->-r requirements_notebook.txt (line 2)) (3.2.5)
    Requirement already satisfied: python-dateutil>=2.7 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from matplotlib>=3.1.1->-r requirements_notebook.txt (line 2)) (2.9.0.post0)
    Requirement already satisfied: llvmlite<0.43,>=0.42.0dev0 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from numba!=0.54.0,<0.60.0,>=0.50.0->alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (0.42.0)
    Requirement already satisfied: pytz>=2020.1 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from pandas<3.0.0,>=1.0.0->alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (2025.2)
    Requirement already satisfied: tzdata>=2022.7 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from pandas<3.0.0,>=1.0.0->alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (2025.2)
    Requirement already satisfied: annotated-types>=0.6.0 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from pydantic<3.0.0,>=1.8.0->alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (0.7.0)
    Requirement already satisfied: pydantic-core==2.41.5 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from pydantic<3.0.0,>=1.8.0->alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (2.41.5)
    Requirement already satisfied: typing-inspection>=0.4.2 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from pydantic<3.0.0,>=1.8.0->alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (0.4.2)
    Requirement already satisfied: six>=1.5 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from python-dateutil>=2.7->matplotlib>=3.1.1->-r requirements_notebook.txt (line 2)) (1.17.0)
    Requirement already satisfied: charset_normalizer<4,>=2 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from requests<3.0.0,>=2.21.0->alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (3.4.4)
    Requirement already satisfied: idna<4,>=2.5 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from requests<3.0.0,>=2.21.0->alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (3.11)
    Requirement already satisfied: urllib3<3,>=1.21.1 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from requests<3.0.0,>=2.21.0->alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (2.6.2)
    Requirement already satisfied: certifi>=2017.4.17 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from requests<3.0.0,>=2.21.0->alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (2025.11.12)
    Requirement already satisfied: networkx>=2.8 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from scikit-image<0.25,>=0.19->alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (3.5)
    Requirement already satisfied: imageio>=2.27 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from scikit-image<0.25,>=0.19->alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (2.37.2)
    Requirement already satisfied: tifffile>=2022.8.12 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from scikit-image<0.25,>=0.19->alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (2025.10.16)
    Requirement already satisfied: lazy_loader>=0.3 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from scikit-image<0.25,>=0.19->alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (0.4)
    Requirement already satisfied: joblib>=1.2.0 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from scikit-learn<2.0.0,>=0.20.2->alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (1.2.0)
    Requirement already satisfied: threadpoolctl>=3.1.0 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from scikit-learn<2.0.0,>=0.20.2->alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (3.6.0)
    Requirement already satisfied: filelock in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from transformers<5.0.0,>=4.0.0->alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (3.20.0)
    Requirement already satisfied: huggingface-hub<1.0,>=0.34.0 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from transformers<5.0.0,>=4.0.0->alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (0.36.0)
    Requirement already satisfied: pyyaml>=5.1 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from transformers<5.0.0,>=4.0.0->alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (6.0.3)
    Requirement already satisfied: regex!=2019.12.17 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from transformers<5.0.0,>=4.0.0->alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (2025.11.3)
    Requirement already satisfied: tokenizers<=0.23.0,>=0.22.0 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from transformers<5.0.0,>=4.0.0->alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (0.22.1)
    Requirement already satisfied: safetensors>=0.4.3 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from transformers<5.0.0,>=4.0.0->alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (0.6.2)
    Requirement already satisfied: fsspec>=2023.5.0 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from huggingface-hub<1.0,>=0.34.0->transformers<5.0.0,>=4.0.0->alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (2025.10.0)
    Requirement already satisfied: hf-xet<2.0.0,>=1.1.3 in /home/tonya/seldon-core/.venv-ad/lib/python3.12/site-packages (from huggingface-hub<1.0,>=0.34.0->transformers<5.0.0,>=4.0.0->alibi-detect>=0.13.0->-r requirements_notebook.txt (line 1)) (1.2.0)


Ensure istio gateway installed


```python
!kubectl apply -f ../../../notebooks/resources/seldon-gateway.yaml
```

    gateway.networking.istio.io/seldon-gateway unchanged



```python
!cat ../../../notebooks/resources/seldon-gateway.yaml
```

    apiVersion: networking.istio.io/v1alpha3
    kind: Gateway
    metadata:
      name: seldon-gateway
      namespace: istio-system
    spec:
      selector:
        istio: ingressgateway # use istio default controller
      servers:
      - port:
          number: 80
          name: http
          protocol: HTTP
        hosts:
        - "*"


## Setup Resources


```python
!kubectl create namespace cifar10
```

    namespace/cifar10 created



```python
%%writefile broker.yaml
apiVersion: eventing.knative.dev/v1
kind: Broker
metadata:
 name: default
 namespace: cifar10
```

    Writing broker.yaml



```python
!kubectl create -f broker.yaml
```

    broker.eventing.knative.dev/default created



```python
%%writefile event-display.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello-display
  namespace: cifar10
spec:
  replicas: 1
  selector:
    matchLabels: &labels
      app: hello-display
  template:
    metadata:
      labels: *labels
    spec:
      containers:
        - name: event-display
          image: gcr.io/knative-releases/knative.dev/eventing/cmd/event_display

---

kind: Service
apiVersion: v1
metadata:
  name: hello-display
  namespace: cifar10
spec:
  selector:
    app: hello-display
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
```

    Writing event-display.yaml



```python
!kubectl apply -f event-display.yaml
```

    deployment.apps/hello-display created
    service/hello-display created


Create the SeldonDeployment image classification model for Cifar10. We add in a `logger` for requests - the default destination is the namespace Knative Broker.


```python
%%writefile cifar10.yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: tfserving-cifar10
  namespace: cifar10
spec:
  protocol: tensorflow
  transport: rest
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - args: 
          - --port=8500
          - --rest_api_port=8501
          - --model_name=resnet32
          - --model_base_path=gs://seldon-models/tfserving/cifar10/resnet32
          image: tensorflow/serving
          name: resnet32
          ports:
          - containerPort: 8501
            name: http
            protocol: TCP
    graph:
      name: resnet32
      type: MODEL
      endpoint:
        service_port: 8501
      logger:
        mode: all
        url: http://broker-ingress.knative-eventing.svc.cluster.local/cifar10/default
    name: model
    replicas: 1

```

    Writing cifar10.yaml



```python
!kubectl apply -f cifar10.yaml
```

    seldondeployment.machinelearning.seldon.io/tfserving-cifar10 created


Create the pretrained VAE Cifar10 Outlier Detector. We forward replies to the message-dumper we started.

Here we configure `seldonio/alibi-detect-server` to use rclone for downloading the artifact. 
If `RCLONE_ENABLED=true` environmental variable is set or any of the environmental variables contain `RCLONE_CONFIG` in their name then rclone
will be used to download the artifacts. If `RCLONE_ENABLED=false` or no `RCLONE_CONFIG` variables are present then kfserving storage.py logic will be used to download the artifacts.


```python
%%writefile cifar10od.yaml

apiVersion: v1
kind: Secret
metadata:
  name: seldon-rclone-secret
  namespace: cifar10    
type: Opaque
stringData:
  RCLONE_CONFIG_GS_TYPE: google cloud storage
  RCLONE_CONFIG_GS_ANONYMOUS: "true"

---            

apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: vae-outlier
  namespace: cifar10
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/minScale: "1"
    spec:
      containers:
      - image: seldonio/alibi-detect-server:1.19.0-dev
        imagePullPolicy: IfNotPresent
        args:
        - --model_name
        - cifar10od
        - --http_port
        - '8080'
        - --protocol
        - tensorflow.http
        - --storage_uri
        - gs://seldon-models/alibi-detect/od/OutlierVAE/cifar10
        - --reply_url
        - http://hello-display.cifar10
        - --event_type
        - io.seldon.serving.inference.outlier
        - --event_source
        - io.seldon.serving.cifar10od
        - OutlierDetector
        envFrom:
        - secretRef:
            name: seldon-rclone-secret
```

    Writing cifar10od.yaml



```python
!kubectl apply -f cifar10od.yaml
```

    secret/seldon-rclone-secret created
    [33;1mWarning:[0m Kubernetes default value is insecure, Knative may default this to secure in a future release: spec.template.spec.containers[0].securityContext.allowPrivilegeEscalation, spec.template.spec.containers[0].securityContext.capabilities, spec.template.spec.containers[0].securityContext.runAsNonRoot, spec.template.spec.containers[0].securityContext.seccompProfile
    service.serving.knative.dev/vae-outlier created


Create a Knative trigger to forward logging events to our Outlier Detector.


```python
%%writefile trigger.yaml
apiVersion: eventing.knative.dev/v1
kind: Trigger
metadata:
  name: vaeoutlier-trigger
  namespace: cifar10
spec:
  broker: default
  filter:
    attributes:
      type: io.seldon.serving.inference.request
  subscriber:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: vae-outlier
      namespace: cifar10

```

    Writing trigger.yaml



```python
!kubectl apply -f trigger.yaml
```

    trigger.eventing.knative.dev/vaeoutlier-trigger created


Get the IP address of the Istio Ingress Gateway. This assumes you have installed istio with a LoadBalancer.


```python
CLUSTER_IPS = !(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
CLUSTER_IP = CLUSTER_IPS[0]
print(CLUSTER_IP)
```

Optionally add an authorization token here if you need one.Acquiring this token will be dependent on your auth setup.


```python
TOKEN = "Bearer <my token>"
```

If you are using Kind or Minikube you will need to port-forward to the istio ingressgateway and uncomment the following


```python
CLUSTER_IP="localhost:8080"
```


```python
SERVICE_HOSTNAMES = !(kubectl get ksvc -n cifar10 vae-outlier -o jsonpath='{.status.url}' | cut -d "/" -f 3)
SERVICE_HOSTNAME_VAEOD = SERVICE_HOSTNAMES[0]
print(SERVICE_HOSTNAME_VAEOD)
```

    vae-outlier.cifar10.svc.cluster.local



```python
import json

import matplotlib.pyplot as plt
import numpy as np
import tensorflow as tf

tf.keras.backend.clear_session()

import requests
from alibi_detect.utils.perturbation import apply_mask

train, test = tf.keras.datasets.cifar10.load_data()
X_train, y_train = train
X_test, y_test = test

X_train = X_train.astype("float32") / 255
X_test = X_test.astype("float32") / 255
print(X_train.shape, y_train.shape, X_test.shape, y_test.shape)
classes = (
    "plane",
    "car",
    "bird",
    "cat",
    "deer",
    "dog",
    "frog",
    "horse",
    "ship",
    "truck",
)


def show(X):
    plt.imshow(X.reshape(32, 32, 3))
    plt.axis("off")
    plt.show()


def predict(X):
    formData = {"instances": X.tolist()}
    headers = {"Authorization": TOKEN}
    res = requests.post(
        "http://"
        + CLUSTER_IP
        + "/seldon/cifar10/tfserving-cifar10/v1/models/resnet32/:predict",
        json=formData,
        headers=headers,
    )
    if res.status_code == 200:
        return classes[np.array(res.json()["predictions"])[0].argmax()]
    else:
        print("Failed with ", res.status_code)
        return []
```

    (50000, 32, 32, 3) (50000, 1) (10000, 32, 32, 3) (10000, 1)


## Normal Prediction


```python
idx = 1
X = X_train[idx : idx + 1]
show(X)
predict(X)
```


    
![png](docs-gb/notebooks/outlier_cifar10_files/docs-gb/notebooks/outlier_cifar10_30_0.png)
    





    'truck'



Lets check the message dumper for an outlier detection prediction. This should be false.


```python
!kubectl logs -n cifar10 $(kubectl get pod -n cifar10 -l app=hello-display -o jsonpath='{.items[0].metadata.name}')
```

    2025/12/12 15:00:08 failed to parse observability config from env, falling back to default config
    2025/12/12 15:00:08 failed to correctly initialize otel resource, resouce may be missing some attributes: the environment variable "SYSTEM_NAMESPACE" is not set, not adding "k8s.namespace.name" to otel attributes
    ☁️  cloudevents.Event
    Context Attributes,
      specversion: 1.0
      type: io.seldon.serving.inference.outlier
      source: io.seldon.serving.cifar10od
      id: 7740b269-da42-4e6f-a451-8ba8a239b97f
    Extensions,
      endpoint: model
      inferenceservicename: tfserving-cifar10
      knativearrivaltime: 2025-12-12T15:03:13.825723818Z
      modelid: resnet32
      namespace: cifar10
      protocol: tensorflow
      requestid: 960f455d-8fee-4cb0-a560-5d2b64870634
      traceparent: 00-4c9e3317bd45a88e22a7ab1a787f47f2-a191b6f624f17efe-00
    Data,
      {"data": {"is_outlier": [0]}, "meta": {"name": "OutlierVAE", "detector_type": "offline", "data_type": "image"}}


## Outlier Prediction


```python
np.random.seed(0)
X_mask, mask = apply_mask(
    X.reshape(1, 32, 32, 3),
    mask_size=(10, 10),
    n_masks=1,
    channels=[0, 1, 2],
    mask_type="normal",
    noise_distr=(0, 1),
    clip_rng=(0, 1),
)
```


```python
show(X_mask)
predict(X_mask)
```


    
![png](docs-gb/notebooks/outlier_cifar10_files/docs-gb/notebooks/outlier_cifar10_35_0.png)
    





    'truck'



Now lets check the message dumper for a new message. This should show we have found an outlier.


```python
!kubectl logs -n cifar10 $(kubectl get pod -n cifar10 -l app=hello-display -o jsonpath='{.items[0].metadata.name}')
```

    2025/12/12 15:00:08 failed to parse observability config from env, falling back to default config
    2025/12/12 15:00:08 failed to correctly initialize otel resource, resouce may be missing some attributes: the environment variable "SYSTEM_NAMESPACE" is not set, not adding "k8s.namespace.name" to otel attributes
    ☁️  cloudevents.Event
    Context Attributes,
      specversion: 1.0
      type: io.seldon.serving.inference.outlier
      source: io.seldon.serving.cifar10od
      id: 7740b269-da42-4e6f-a451-8ba8a239b97f
    Extensions,
      endpoint: model
      inferenceservicename: tfserving-cifar10
      knativearrivaltime: 2025-12-12T15:03:13.825723818Z
      modelid: resnet32
      namespace: cifar10
      protocol: tensorflow
      requestid: 960f455d-8fee-4cb0-a560-5d2b64870634
      traceparent: 00-4c9e3317bd45a88e22a7ab1a787f47f2-a191b6f624f17efe-00
    Data,
      {"data": {"is_outlier": [0]}, "meta": {"name": "OutlierVAE", "detector_type": "offline", "data_type": "image"}}
    ☁️  cloudevents.Event
    Context Attributes,
      specversion: 1.0
      type: io.seldon.serving.inference.outlier
      source: io.seldon.serving.cifar10od
      id: 2337a67e-ced2-4807-a907-938e4ecea304
    Extensions,
      endpoint: model
      inferenceservicename: tfserving-cifar10
      knativearrivaltime: 2025-12-12T15:03:35.060332451Z
      modelid: resnet32
      namespace: cifar10
      protocol: tensorflow
      requestid: 892e75f7-01b4-4690-be81-ae7a706679b2
      traceparent: 00-8b8b0df29bc049f0a0d766f1fcc6a667-8d2e8aa417b2dfc6-00
    Data,
      {"data": {"is_outlier": [1]}, "meta": {"name": "OutlierVAE", "detector_type": "offline", "data_type": "image"}}


## Tear Down


```python
!kubectl delete ns cifar10
```

    namespace "cifar10" deleted

