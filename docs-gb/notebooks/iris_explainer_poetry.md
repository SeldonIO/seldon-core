# Explainer for Iris model with Poetry-defined Environment

## Prerequisites

 * A kubernetes cluster with kubectl configured
 * poetry
 * rclone
 * curl

## Setup Seldon Core

Use the setup notebook to [Setup Cluster](../notebooks/seldon-core-setup.md#setup-cluster) with [Ambassador Ingress](../notebooks/seldon-core-setup.md#ambassador) and [Install Seldon Core](../notebooks/seldon-core-setup.md#Install-Seldon-Core). Instructions [also online](../notebooks/seldon-core-setup.md).

We will assume that ambassador (or Istio) ingress is port-forwarded to `localhost:8003`

## Setup MinIO

Use the provided [notebook](../notebooks/minio_setup.md) to install Minio in your cluster.
Instructions [also online](../notebooks/minio_setup.md).

We will assume that MinIO service is port-forwarded to `localhost:8090`


```python
%%writefile rclone.conf
[s3]
type = s3
provider = minio
env_auth = false
access_key_id = minioadmin
secret_access_key = minioadmin
endpoint = http://localhost:8090
```

    Overwriting rclone.conf



```python
%%writefile secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: seldon-rclone-secret
type: Opaque
stringData:
  RCLONE_CONFIG_S3_TYPE: s3
  RCLONE_CONFIG_S3_PROVIDER: minio
  RCLONE_CONFIG_S3_ENV_AUTH: "false"
  RCLONE_CONFIG_S3_ACCESS_KEY_ID: minioadmin
  RCLONE_CONFIG_S3_SECRET_ACCESS_KEY: minioadmin
  RCLONE_CONFIG_S3_ENDPOINT: http://minio.minio-system.svc.cluster.local:9000
```

    Overwriting secret.yaml



```python
!kubectl apply -f secret.yaml
```

    secret/seldon-rclone-secret configured


## Poetry

We will use `poetry.lock` to fully define the explainer environment. Install poetry following official [documentation](https://python-poetry.org/docs/#installation). Usually this goes down to
```bash
curl -sSL https://install.python-poetry.org | python3 - --version 1.1.15
```

# Deploy Iris Model


```python
%%writefile iris.yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: iris
spec:
  predictors:
  - name: default
    replicas: 1
    graph:
      name: classifier
      implementation: SKLEARN_SERVER
      modelUri: gs://seldon-models/v1.15.0-dev/sklearn/iris
```

    Overwriting iris.yaml



```python
!kubectl apply -f iris.yaml
```

    seldondeployment.machinelearning.seldon.io/iris configured



```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=iris -o jsonpath='{.items[0].metadata.name}')
```

    Waiting for deployment "iris-default-0-classifier" rollout to finish: 1 old replicas are pending termination...
    Waiting for deployment "iris-default-0-classifier" rollout to finish: 1 old replicas are pending termination...
    deployment "iris-default-0-classifier" successfully rolled out



```bash
%%bash
curl -s -X POST -H 'Content-Type: application/json' \
    -d '{"data":{"ndarray":[[5.964, 4.006, 2.081, 1.031]]}}' \
    http://localhost:8003/seldon/seldon/iris/api/v1.0/predictions  | jq .
```

    {
      "data": {
        "names": [
          "t:0",
          "t:1",
          "t:2"
        ],
        "ndarray": [
          [
            0.9548873249364059,
            0.04505474761562512,
            5.7927447968953825e-05
          ]
        ]
      },
      "meta": {
        "requestPath": {
          "classifier": "seldonio/sklearnserver:1.15.0-dev"
        }
      }
    }


# Train Explainer

## Prepare Training Environment

We are going to use `pyproject.toml` and `poetry.lock` files from [Alibi Explain Server](https://github.com/SeldonIO/seldon-core/tree/master/components/alibi-explain-server). This will allow us to create environment that will match the runtime one.


```bash
%%bash
cp ../../../components/alibi-explain-server/pyproject.toml .
cp ../../../components/alibi-explain-server/poetry.lock .

conda create --yes --prefix ./venv python=3.7.10
```

    Collecting package metadata (current_repodata.json): ...working... done
    Solving environment: ...working... failed with repodata from current_repodata.json, will retry with next repodata source.
    Collecting package metadata (repodata.json): ...working... done
    Solving environment: ...working... done
    
    ## Package Plan ##
    
      environment location: /home/rskolasinski/work/seldon-core/examples/explainers/iris-explainer-poetry/venv
    
      added / updated specs:
        - conda-ecosystem-user-package-isolation
        - python=3.7.10
    
    
    The following NEW packages will be INSTALLED:
    
      _libgcc_mutex      conda-forge/linux-64::_libgcc_mutex-0.1-conda_forge
      _openmp_mutex      conda-forge/linux-64::_openmp_mutex-4.5-2_gnu
      ca-certificates    conda-forge/linux-64::ca-certificates-2022.5.18.1-ha878542_0
      conda-ecosystem-u~ conda-forge/linux-64::conda-ecosystem-user-package-isolation-1.0-ha770c72_1
      ld_impl_linux-64   conda-forge/linux-64::ld_impl_linux-64-2.36.1-hea4e1c9_2
      libffi             conda-forge/linux-64::libffi-3.4.2-h7f98852_5
      libgcc-ng          conda-forge/linux-64::libgcc-ng-12.1.0-h8d9b700_16
      libgomp            conda-forge/linux-64::libgomp-12.1.0-h8d9b700_16
      libnsl             conda-forge/linux-64::libnsl-2.0.0-h7f98852_0
      libstdcxx-ng       conda-forge/linux-64::libstdcxx-ng-12.1.0-ha89aaad_16
      libzlib            conda-forge/linux-64::libzlib-1.2.11-h166bdaf_1014
      ncurses            conda-forge/linux-64::ncurses-6.3-h27087fc_1
      openssl            conda-forge/linux-64::openssl-3.0.3-h166bdaf_0
      pip                conda-forge/noarch::pip-22.1.1-pyhd8ed1ab_0
      python             conda-forge/linux-64::python-3.7.10-hf930737_104_cpython
      python_abi         conda-forge/linux-64::python_abi-3.7-2_cp37m
      readline           conda-forge/linux-64::readline-8.1-h46c0cb4_0
      setuptools         conda-forge/linux-64::setuptools-62.3.2-py37h89c1867_0
      sqlite             conda-forge/linux-64::sqlite-3.38.5-h4ff8645_0
      tk                 conda-forge/linux-64::tk-8.6.12-h27826a3_0
      wheel              conda-forge/noarch::wheel-0.37.1-pyhd8ed1ab_0
      xz                 conda-forge/linux-64::xz-5.2.5-h516909a_1
      zlib               conda-forge/linux-64::zlib-1.2.11-h166bdaf_1014
    
    
    Preparing transaction: ...working... done
    Verifying transaction: ...working... done
    Executing transaction: ...working... done
    #
    # To activate this environment, use
    #
    #     $ conda activate /home/rskolasinski/work/seldon-core/examples/explainers/iris-explainer-poetry/venv
    #
    # To deactivate an active environment, use
    #
    #     $ conda deactivate
    



```bash
%%bash
source ~/miniconda3/etc/profile.d/conda.sh
conda activate ./venv
poetry install
```

    Installing dependencies from lock file
    
    Package operations: 138 installs, 0 updates, 0 removals
    
      • Installing certifi (2021.10.8)
      • Installing charset-normalizer (2.0.11)
      • Installing idna (3.3)
      • Installing pyasn1 (0.4.8)
      • Installing pycparser (2.21)
      • Installing typing-extensions (4.0.1)
      • Installing urllib3 (1.26.8)
      • Installing zipp (3.7.0)
      • Installing cachetools (5.0.0)
      • Installing cffi (1.15.0)
      • Installing cymem (2.0.6)
      • Installing importlib-metadata (4.10.1)
      • Installing numpy (1.19.5)
      • Installing oauthlib (3.2.0)
      • Installing pyparsing (3.0.7)
      • Installing requests (2.27.1)
      • Installing rsa (4.7.2)
      • Installing six (1.16.0)
      • Installing pyasn1-modules (0.2.8)
      • Installing murmurhash (1.0.6)
      • Installing blis (0.7.5)
      • Installing catalogue (1.0.0)
      • Installing click (8.0.3)
      • Installing cryptography (36.0.1)
      • Installing filelock (3.4.2)
      • Installing google-auth (2.6.0)
      • Installing joblib (1.1.0)
      • Installing packaging (21.3)
      • Installing pillow (9.0.0)
      • Installing preshed (3.0.6)
      • Installing httplib2 (0.20.2)
      • Installing pyyaml (6.0)
      • Installing pyu2f (0.1.5)
      • Installing plac (1.1.3)
      • Installing requests-oauthlib (1.3.1)
      • Installing regex (2022.1.18)
      • Installing sniffio (1.2.0)
      • Installing srsly (1.0.5)
      • Installing tomli (1.2.3)
      • Installing tqdm (4.62.3)
      • Installing wasabi (0.9.0)
      • Installing absl-py (1.0.0)
      • Installing anyio (3.5.0)
      • Installing attrs (21.4.0)
      • Installing boto (2.49.0)
      • Installing cached-property (1.5.2)
      • Installing cycler (0.11.0)
      • Installing fasteners (0.17.3)
      • Installing fonttools (4.29.1)
      • Installing google-auth-oauthlib (0.4.6)
      • Installing google-reauth (0.1.1)
      • Installing grpcio (1.43.0)
      • Installing huggingface-hub (0.4.0)
      • Installing imageio (2.14.1)
      • Installing iniconfig (1.1.1)
      • Installing kiwisolver (1.3.2)
      • Installing llvmlite (0.38.0)
      • Installing markdown (3.3.6)
      • Installing networkx (2.6.3)
      • Installing oauth2client (4.1.3)
      • Installing pluggy (1.0.0)
      • Installing protobuf (3.19.4)
      • Installing py (1.11.0)
      • Installing python-dateutil (2.8.2)
      • Installing pyopenssl (22.0.0)
      • Installing pytz (2021.3)
      • Installing pywavelets (1.2.0)
      • Installing retry-decorator (1.1.1)
      • Installing sacremoses (0.0.47)
      • Installing scipy (1.7.3)
      • Installing setuptools-scm (6.4.2)
      • Installing spacy-lookups-data (0.3.2)
      • Installing tenacity (8.0.1)
      • Installing tensorboard-data-server (0.6.1)
      • Installing tensorboard-plugin-wit (1.8.1)
      • Installing thinc (7.4.5)
      • Installing threadpoolctl (3.1.0)
      • Installing tifffile (2021.11.2)
      • Installing tokenizers (0.11.4)
      • Installing toml (0.10.2)
      • Installing werkzeug (2.0.2)
      • Installing appdirs (1.4.4)
      • Installing argcomplete (2.0.0)
      • Installing astunparse (1.6.3)
      • Installing cloudpickle (2.0.0)
      • Installing crcmod (1.7)
      • Installing dill (0.3.4)
      • Installing gast (0.4.0)
      • Installing gcs-oauth2-boto-plugin (3.0)
      • Installing google-pasta (0.2.0)
      • Installing graphviz (0.19.1)
      • Installing h5py (3.6.0)
      • Installing flatbuffers (2.0)
      • Installing keras-preprocessing (1.1.2)
      • Installing keras (2.7.0)
      • Installing libclang (13.0.0)
      • Installing google-apitools (0.5.32)
      • Installing matplotlib (3.5.1)
      • Installing monotonic (1.6)
      • Installing mypy-extensions (0.4.3)
      • Installing numba (0.55.0)
      • Installing opt-einsum (3.3.0)
      • Installing pandas (1.1.5)
      • Installing pathspec (0.9.0)
      • Installing plotly (5.5.0)
      • Installing ptable (0.9.2)
      • Installing pydantic (1.9.0)
      • Installing pytest (6.2.5)
      • Installing scikit-image (0.19.1)
      • Installing scikit-learn (1.0.2)
      • Installing slicer (0.0.7)
      • Installing spacy (2.3.7)
      • Installing starlette (0.17.1)
      • Installing tensorboard (2.8.0)
      • Installing tensorflow-estimator (2.7.0)
      • Installing tensorflow-io-gcs-filesystem (0.23.1)
      • Installing termcolor (1.1.0)
      • Installing tornado (6.1)
      • Installing transformers (4.16.2)
      • Installing typed-ast (1.4.3)
      • Installing wrapt (1.13.3)
      • Installing alibi (0.7.0)
      • Installing black (21.7b0)
      • Installing catboost (1.0.4)
      • Installing fastapi (0.73.0)
      • Installing grpcio-tools (1.31.0)
      • Installing gsutil (5.5)
      • Installing isort (5.9.0)
      • Installing lightgbm (3.3.2)
      • Installing mypy (0.910)
      • Installing mypy-protobuf (1.22)
      • Installing pip-licenses (3.5.3)
      • Installing pytest-tornasync (0.6.0.post2)
      • Installing requests-mock (1.9.3)
      • Installing shap (0.40.0 429fb3e)
      • Installing tensorflow (2.7.0)
      • Installing types-requests (2.26.0)
      • Installing xgboost (1.5.2)


## Prepare Training Script


```python
%%writefile train.py
import numpy as np
from sklearn.datasets import load_iris
from alibi.explainers import AnchorTabular

import requests


dataset = load_iris()
feature_names = dataset.feature_names
iris_data = dataset.data

model_url = "http://localhost:8003/seldon/seldon/iris/api/v1.0/predictions"


def predict_fn(X):
    data = {"data": {"ndarray": X.tolist()}}
    r = requests.post(model_url, json={"data": {"ndarray": [[1, 2, 3, 4]]}})
    return np.array(r.json()["data"]["ndarray"])


explainer = AnchorTabular(predict_fn, feature_names)
explainer.fit(iris_data, disc_perc=(25, 50, 75))

explainer.save("./explainer/")
```

    Overwriting train.py



```bash
%%bash
unset MPLBACKEND # required as we call the script from Jupyter Lab in this demo
./venv/bin/python3 train.py
```

    IPython could not be loaded!



```python
!tree explainer/
```

    [01;34mexplainer/[0m
    ├── [00mexplainer.dill[0m
    └── [00mmeta.dill[0m
    
    0 directories, 2 files


# Save and deploy Explainer


```python
!rclone --config="rclone.conf" copy explainer/ s3:explainers/iris/
```


```python
%%writefile iris-with-explainer.yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: iris
spec:
  predictors:
  - name: default
    replicas: 1
    graph:
      name: classifier
      implementation: SKLEARN_SERVER
      modelUri: gs://seldon-models/v1.11.0-dev/sklearn/iris
    explainer:
      type: AnchorTabular
      modelUri: s3:explainers/iris/
      envSecretRefName: seldon-rclone-secret
      replicas: 1            
```

    Overwriting iris-with-explainer.yaml



```python
!kubectl apply -f iris-with-explainer.yaml
```

    seldondeployment.machinelearning.seldon.io/iris configured



```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=iris -o jsonpath='{.items[0].metadata.name}')
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=iris -o jsonpath='{.items[1].metadata.name}')
```

    Waiting for deployment "iris-default-0-classifier" rollout to finish: 1 old replicas are pending termination...
    Waiting for deployment "iris-default-0-classifier" rollout to finish: 1 old replicas are pending termination...
    deployment "iris-default-0-classifier" successfully rolled out
    deployment "iris-default-explainer" successfully rolled out


# Test Deployed explainer


```python
import numpy as np
import requests
```


```python
model_url = (
    "http://localhost:8003/seldon/seldon/iris-explainer/default/api/v1.0/explain"
)


def explain_fn(X):
    data = {"data": {"ndarray": X.tolist()}}
    r = requests.post(model_url, json={"data": {"ndarray": [[1, 2, 3, 4]]}})
    return r.json()
```


```python
explanation = explain_fn(np.array([[5.964, 4.006, 2.081, 1.031]]))
```


```python
print("Anchor: %s" % (" AND ".join(explanation["data"]["anchor"])))
print("Precision: %.2f" % explanation["data"]["precision"])
print("Coverage: %.2f" % explanation["data"]["coverage"])
```

    Anchor: petal width (cm) > 1.80 AND sepal width (cm) <= 2.80
    Precision: 0.97
    Coverage: 0.08

