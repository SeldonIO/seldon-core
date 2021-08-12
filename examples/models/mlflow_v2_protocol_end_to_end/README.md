# MLflow v2 protocol elasticnet wine example

In this example we are going to build a model using mlflow, pack and deploy it on seldon-core on a local kind cluster

Prerequisites before running this notebook:

- install and configure `mc`, follow the relevant section in this [link](https://docs.seldon.io/projects/seldon-core/en/latest/examples/minio_setup.html)

- run this jupyter notebook in conda environment
```bash
$ conda create --name python3.8-mlflow-example python=3.8 -y
$ conda activate python3.8-mlflow-example
$ pip install jupyter
$ jupyter notebook
```

## Setup `seldon-core` and `minio`


### Setup Seldon Core

Use the setup notebook to [Setup Cluster](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Setup-Cluster) with [Ambassador Ingress](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Ambassador) and [Install Seldon Core](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Install-Seldon-Core). Instructions [also online](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html).

### Setup MinIO

Use the provided [notebook](https://docs.seldon.io/projects/seldon-core/en/latest/examples/minio_setup.html) to install Minio in your cluster and configure `mc` CLI tool. 
Instructions [also online](https://docs.seldon.io/projects/seldon-core/en/latest/examples/minio_setup.html).

## Alternatively setup `seldon-core` and `minio` using `ansible`

### Install `ansible` prerequisites


```python
!pip install ansible openshift docker passlib
```


```python
!ansible-galaxy collection install git+https://github.com/SeldonIO/ansible-k8s-collection.git
```

### Setup kind cluster and install `seldon-core`


```bash
%%bash
cd ../../../ansible
ansible-playbook playbooks/kind_cluster.yaml
ansible-playbook playbooks/main.yaml -e seldon_core_version=master 
```

## Train elasticnet wine model using `mlflow`

### Install `mlflow` and required dependencies to train the model


```python
!pip install mlflow scikit-learn==0.23.2 pandas
```

### Define where the model artifacts will be saved


```python
import os
from pathlib import Path
MODEL_DIR = Path(os.getcwd()) / "elasticnet_wine_model"
```

### Define training function


```python
# Wine Quality Sample a copy from: 
# https://github.com/mlflow/mlflow/blob/master/examples/sklearn_elasticnet_wine/train.ipynb

def train(in_alpha, in_l1_ratio):
    import os
    import warnings
    import sys

    import pandas as pd
    import numpy as np
    from sklearn.metrics import mean_squared_error, mean_absolute_error, r2_score
    from sklearn.model_selection import train_test_split
    from sklearn.linear_model import ElasticNet

    import mlflow
    import mlflow.sklearn
    
    import logging
    logging.basicConfig(level=logging.WARN)
    logger = logging.getLogger(__name__)

    def eval_metrics(actual, pred):
        rmse = np.sqrt(mean_squared_error(actual, pred))
        mae = mean_absolute_error(actual, pred)
        r2 = r2_score(actual, pred)
        return rmse, mae, r2


    warnings.filterwarnings("ignore")
    np.random.seed(40)

    # Read the wine-quality csv file from the URL
    csv_url =\
        'http://archive.ics.uci.edu/ml/machine-learning-databases/wine-quality/winequality-red.csv'
    try:
        data = pd.read_csv(csv_url, sep=';')
    except Exception as e:
        logger.exception(
            "Unable to download training & test CSV, check your internet connection. Error: %s", e)

    # Split the data into training and test sets. (0.75, 0.25) split.
    train, test = train_test_split(data)

    # The predicted column is "quality" which is a scalar from [3, 9]
    train_x = train.drop(["quality"], axis=1)
    test_x = test.drop(["quality"], axis=1)
    train_y = train[["quality"]]
    test_y = test[["quality"]]

    # Set default values if no alpha is provided
    if float(in_alpha) is None:
        alpha = 0.5
    else:
        alpha = float(in_alpha)

    # Set default values if no l1_ratio is provided
    if float(in_l1_ratio) is None:
        l1_ratio = 0.5
    else:
        l1_ratio = float(in_l1_ratio)

    # Useful for multiple runs (only doing one run in this sample notebook)    
    with mlflow.start_run():
        # Execute ElasticNet
        lr = ElasticNet(alpha=alpha, l1_ratio=l1_ratio, random_state=42)
        lr.fit(train_x, train_y)

        # Evaluate Metrics
        predicted_qualities = lr.predict(test_x)
        (rmse, mae, r2) = eval_metrics(test_y, predicted_qualities)

        # Print out metrics
        print("Elasticnet model (alpha=%f, l1_ratio=%f):" % (alpha, l1_ratio))
        print("  RMSE: %s" % rmse)
        print("  MAE: %s" % mae)
        print("  R2: %s" % r2)

        # Log parameter, metrics, and model to MLflow
        mlflow.log_param("alpha", alpha)
        mlflow.log_param("l1_ratio", l1_ratio)
        mlflow.log_metric("rmse", rmse)
        mlflow.log_metric("r2", r2)
        mlflow.log_metric("mae", mae)

        mlflow.sklearn.save_model(lr, MODEL_DIR)
        print(f" Model saved to {MODEL_DIR}")
```

### Train the elasticnet_wine model


```python
train(0.5, 0.5)
```

### Install dependencies to be able to pack and deploy the model on `seldon_core`

We are going to use [`conda-pack`](https://conda.github.io/conda-pack/) to pack the python enviornment. We also need `mlserver` dependencies.
We are planning to simplify this workflow in future releases.


```python
!pip install conda-pack mlserver==0.4.0 mlserver-mlflow==0.4.0
```

### Pack the conda enviornment


```python
import conda_pack
env_file_path = MODEL_DIR / "environment.tar.gz"
conda_pack.pack(
    output=str(env_file_path),
    force=True,
    verbose=True,
    ignore_editable_packages=False,
    ignore_missing_files=True,
)
```

### Configure `mc` to access the minio service in the local kind cluster
note: make sure that minio ip is reflected properly below, run `kubectl get service -n minio-system`


```python
!mc config host add minio-seldon http://172.18.255.3:9000 minioadmin minioadmin
```

### Copy the model artifacts to minio


```python
import os
target_bucket = "minio-seldon/model"
os.system(f"mc rb --force {target_bucket}")
os.system(f"mc mb {target_bucket}")
os.system(f"mc cp --recursive {MODEL_DIR} {target_bucket}")
```

### Create model deployment configuration


```python
%%writefile mlflow_elasticnet_wine_v2.yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: mlflow
spec:
  protocol: kfserving  # Activate v2 protocol
  name: wines
  predictors:
    - graph:
        children: []
        implementation: MLFLOW_SERVER
        modelUri: s3:/model
        envSecretRefName: seldon-rclone-secret
        name: classifier
      name: default
      replicas: 1
```

### Deploy the model on the local kind cluster


```python
!kubectl apply -f mlflow_elasticnet_wine_v2.yaml
```


```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=mlflow -o jsonpath='{.items[0].metadata.name}')
```

### Get prediction from the service using REST


```python
import json

import requests

inference_request = {
    "parameters": {
        "content_type": "pd"
    },
    "inputs": [
        {
          "name": "fixed acidity",
          "shape": [1],
          "datatype": "FP32",
          "data": [7.4],
          "parameters": {
              "content_type": "np"
          }
        },
        {
          "name": "volatile acidity",
          "shape": [1],
          "datatype": "FP32",
          "data": [0.7000],
          "parameters": {
              "content_type": "np"
          }
        },
        {
          "name": "citric acidity",
          "shape": [1],
          "datatype": "FP32",
          "data": [0],
          "parameters": {
              "content_type": "np"
          }
        },
        {
          "name": "residual sugar",
          "shape": [1],
          "datatype": "FP32",
          "data": [1.9],
          "parameters": {
              "content_type": "np"
          }
        },
        {
          "name": "chlorides",
          "shape": [1],
          "datatype": "FP32",
          "data": [0.076],
          "parameters": {
              "content_type": "np"
          }
        },
        {
          "name": "free sulfur dioxide",
          "shape": [1],
          "datatype": "FP32",
          "data": [11],
          "parameters": {
              "content_type": "np"
          }
        },
        {
          "name": "total sulfur dioxide",
          "shape": [1],
          "datatype": "FP32",
          "data": [34],
          "parameters": {
              "content_type": "np"
          }
        },
        {
          "name": "density",
          "shape": [1],
          "datatype": "FP32",
          "data": [0.9978],
          "parameters": {
              "content_type": "np"
          }
        },
        {
          "name": "pH",
          "shape": [1],
          "datatype": "FP32",
          "data": [3.51],
          "parameters": {
              "content_type": "np"
          }
        },
        {
          "name": "sulphates",
          "shape": [1],
          "datatype": "FP32",
          "data": [0.56],
          "parameters": {
              "content_type": "np"
          }
        },
        {
          "name": "alcohol",
          "shape": [1],
          "datatype": "FP32",
          "data": [9.4],
          "parameters": {
              "content_type": "np"
          }
        },
    ]
}

# note is the local balancer for istion, make sure that the ip is reflected in the setup,
# run kubectl get service -n istio-system
endpoint = "http://172.18.255.1/seldon/seldon/mlflow/v2/models/infer"
response = requests.post(endpoint, json=inference_request)

print(json.dumps(response.json(), indent=2))
assert response.ok
```

### Delete the model deployment


```python
!kubectl delete -f mlflow_elasticnet_wine_v2.yaml
```
