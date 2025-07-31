# Upgrade to rclone-based Storage Initializer - automation for AWS S3 / MinIO configuration

In this documentation page we provide an example upgrade path from kfserving-based to rclone-based storage initializer. This is required due to the fact that secret format expected by these two storage initializers is different. 

Storage initializers are used by Seldon's pre-packaged model servers to download models binaries. 
As it is explained in the [SC 1.8 upgrading notes](../upgrading.md#upgrading-to-1-8) the [seldonio/rclone-storage-initializer](https://github.com/SeldonIO/seldon-core/tree/master/components/rclone-storage-initializer) became default storage initializer in v1.8.0.

In this tutorial we aim to provide an intuition of the steps you will have to carry to migrate to the new rclone-based Storage Initializer with the context that every cluster configuration will be different, so you should be able to see this as something you can build from.

Read more:
- [Prepackaged Model Servers documentation page](../servers/overview.md)
- [SC 1.8 upgrading notes](../upgrading.md#upgrading-to-1-8)
- [Testing new storage initializer without global update](../notebooks/rclone-upgrade.md)

## Prerequisites

 * A kubernetes cluster with kubectl configured
 * mc client
 * curl

## Steps in this tutorial
 
 * Start with SC configured to use kfserving-based storage initializer
 * Copy iris model from GCS into in-cluster minio
 * Deploy SKlearn Pre-Packaged server using kfserving storage initializer
     * Providing credentials using old-style storage initializer secret
     * Providing credentials using old-style storage initializer Service Account format
 * Extend secrets to include rclone-specific fields (patch Seldon Deployments where required)
 * Upgrade SC installation to use rclone-based storage initializer
 
## Setup Seldon Core

Use the setup notebook to [Setup Cluster](../notebooks/seldon-core-setup.md#setup-cluster) with [Ambassador Ingress](../notebooks/seldon-core-setup.md#ambassador) and [Install Seldon Core](../notebooks/seldon-core-setup.md#Install-Seldon-Core). 

Set starting storage initializer to be kfserving one


```bash
%%bash
helm upgrade seldon-core seldon-core-operator \
    --install \
    --repo https://storage.googleapis.com/seldon-charts \
    --version 1.9.1 \
    --namespace seldon-system \
    --set storageInitializer.image="kfserving/storage-initializer:v0.6.1" \
    --reuse-values
```

## Setup MinIO

Use the provided [notebook](../notebooks/minio_setup.md) to install Minio in your cluster and configure `mc` CLI tool. 

## Copy iris model into local MinIO


```bash
%%bash
mc config host add gcs https://storage.googleapis.com "" "" 

mc mb minio-seldon/sklearn/iris/ -p
mc cp gcs/seldon-models/sklearn/iris/model.joblib minio-seldon/sklearn/iris/
mc cp gcs/seldon-models/sklearn/iris/metadata.yaml minio-seldon/sklearn/iris/
```


```bash
%%bash
mc ls minio-seldon/sklearn/iris/
```

## Deploy SKLearn Server with kfserving-storage-initializer

First we deploy the model using kfserving-storage-initializer. This is using the default Storage Initializer for pre Seldon Core v1.8.0.

## Using envSecretRefName


```python
%%writefile sklearn-iris-secret.yaml

apiVersion: v1
kind: Secret
metadata:
  name: seldon-kfserving-secret
type: Opaque
stringData:
  AWS_ACCESS_KEY_ID: minioadmin
  AWS_SECRET_ACCESS_KEY: minioadmin
  AWS_ENDPOINT_URL: http://minio.minio-system.svc.cluster.local:9000
  USE_SSL: "false"
    
---
    
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: sklearn-iris-secret
spec:
  predictors:
  - name: default
    replicas: 1
    graph:
      name: classifier
      implementation: SKLEARN_SERVER
      modelUri: s3://sklearn/iris
      envSecretRefName: seldon-kfserving-secret
```


```python
!kubectl apply -f sklearn-iris-secret.yaml
```


```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=sklearn-iris-secret -o jsonpath='{.items[0].metadata.name}')
```


```bash
%%bash
curl -s -X POST -H 'Content-Type: application/json' \
    -d '{"data":{"ndarray":[[5.964, 4.006, 2.081, 1.031]]}}' \
    http://localhost:8003/seldon/seldon/sklearn-iris-secret/api/v1.0/predictions  | jq .
```

## Using serviceAccountName


```python
%%writefile sklearn-iris-sa.yaml

apiVersion: v1
kind: ServiceAccount
metadata:
  name: minio-sa
secrets:
  - name: minio-sa-secret

---

apiVersion: v1
kind: Secret
metadata:
  name: minio-sa-secret
  annotations:
     machinelearning.seldon.io/s3-endpoint: minio.minio-system.svc.cluster.local:9000
     machinelearning.seldon.io/s3-usehttps: "0"
type: Opaque
stringData:
  awsAccessKeyID: "minioadmin"
  awsSecretAccessKey: "minioadmin"

---
    
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: sklearn-iris-sa
spec:
  predictors:
  - name: default
    replicas: 1
    graph:
      name: classifier
      implementation: SKLEARN_SERVER
      modelUri: s3://sklearn/iris
      serviceAccountName: minio-sa
```


```python
!kubectl apply -f sklearn-iris-sa.yaml
```


```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=sklearn-iris-sa -o jsonpath='{.items[0].metadata.name}')
```


```bash
%%bash
curl -s -X POST -H 'Content-Type: application/json' \
    -d '{"data":{"ndarray":[[5.964, 4.006, 2.081, 1.031]]}}' \
    http://localhost:8003/seldon/seldon/sklearn-iris-sa/api/v1.0/predictions  | jq .
```

## Preparing rclone-compatible secret

The [rclone](https://rclone.org/)-based storage initializer expects one to define a new secret. General documentation credentials hadling can be found [here](../servers/overview.md#handling-credentials) with constantly updated examples of tested configurations.

If we do not have yet an example for Cloud Storage solution that you are using, please, consult the relevant page on [RClone documentation](https://rclone.org/#providers).

## Updating envSecretRefName-specified secrets


```python
from typing import Dict, List, Tuple, Union

from kubernetes import client, config

AWS_SECRET_REQUIRED_FIELDS = [
    "AWS_ACCESS_KEY_ID",
    "AWS_ENDPOINT_URL",
    "AWS_SECRET_ACCESS_KEY",
]


def get_secrets_to_update(namespace: str) -> List[str]:
    """Get list of secrets defined for Seldon Deployments in a given namespace.

    Parameters:
    ----------
    namespace: str
        Namespace in which to look for secrets attached to Seldon Deployments.

    Returns:
    -------
    secrets_names: List[str]
        List of secrets names
    """
    secret_names = []
    api_instance = client.CustomObjectsApi()
    sdeps = api_instance.list_namespaced_custom_object(
        "machinelearning.seldon.io",
        "v1",
        namespace,
        "seldondeployments",
    )
    for sdep in sdeps.get("items", []):
        for predictor in sdep.get("spec", {}).get("predictors", []):
            secret_name = predictor.get("graph", {}).get("envSecretRefName", None)
            if secret_name:
                secret_names.append(secret_name)
    return secret_names


def new_fields_for_secret(secret: client.V1Secret, provider: str) -> Dict:
    """Get new fields that need to be added to secret.

    Parameters
    ----------
    secret: client.V1Secret
        Kubernetes secret that needs to be updated
    provider: str
        S3 provider: must be minio or aws

    Returns
    -------
    new_fields: dict
        New fields for  the secret partitioned into 'data' and 'stringData' fields
    """
    for key in AWS_SECRET_REQUIRED_FIELDS:
        if key not in secret.data:
            raise ValueError(
                f"Secret '{secret.metadata.name}' does not contain '{key}' field."
            )

    return {
        "data": {
            "RCLONE_CONFIG_S3_ACCESS_KEY_ID": secret.data.get("AWS_ACCESS_KEY_ID"),
            "RCLONE_CONFIG_S3_SECRET_ACCESS_KEY": secret.data.get(
                "AWS_SECRET_ACCESS_KEY"
            ),
            "RCLONE_CONFIG_S3_ENDPOINT": secret.data.get("AWS_ENDPOINT_URL"),
        },
        "stringData": {
            "RCLONE_CONFIG_S3_TYPE": "s3",
            "RCLONE_CONFIG_S3_PROVIDER": provider,
            "RCLONE_CONFIG_S3_ENV_AUTH": "false",
        },
    }


def update_aws_secrets(namespaces: List[str], provider: str):
    """Updated AWS secrets used by Seldon Deployments in specified namespaces

    Parameters
    ----------
    namespaces: List[str]
        List of namespaces in which will look for Seldon Deployments
    provider: str
        S3 provider: must be minio or aws
    """
    if provider not in ["minio", "aws"]:
        raise ValueError("Provider must be 'minio' or 'aws'")

    v1 = client.CoreV1Api()
    for namespace in namespaces:
        print(f"Updating secrets in namespace {namespace}")
        secret_names = get_secrets_to_update(namespace)
        for secret_name in secret_names:
            secret = v1.read_namespaced_secret(secret_name, namespace)
            try:
                new_fields = new_fields_for_secret(secret, provider)
            except ValueError as e:
                print(f"  Couldn't upgrade a secret: {e}.")
                continue
            _ = v1.patch_namespaced_secret(
                secret_name,
                namespace,
                client.V1Secret(
                    data=new_fields["data"], string_data=new_fields["stringData"]
                ),
            )
            print(f"  Upgraded secret {secret_name}.")
```


```python
config.load_kube_config()
update_aws_secrets(namespaces=["seldon"], provider="minio")
```

### Updating serviceAccountName-specified secrets and deployments


```python
AWS_SA_SECRET_REQUIRED_FIELDS = ["awsAccessKeyID", "awsSecretAccessKey"]

AWS_SA_SECRET_REQUIRED_ANNOTATIONS = [
    "machinelearning.seldon.io/s3-usehttps",
    "machinelearning.seldon.io/s3-endpoint",
]


def get_sdeps_with_service_accounts(namespace: str) -> List[Tuple[dict, List[str]]]:
    """Get list of secrets defined for Seldon Deployments in a given namespace.

    Parameters:
    ----------
    namespace: str
        Namespace in which to look for secrets attached to Seldon Deployments.

    Returns:
    -------
    output: List[Tuple[dict, List[dict]]]]
        Eeach tuple contain sdep (dict) and a list service account names (List[str])
        The list of Service Account names is of length of number of predictors.
        If Predictor has no related Service Account a None is included.
    """
    output = []
    api_instance = client.CustomObjectsApi()
    sdeps = api_instance.list_namespaced_custom_object(
        "machinelearning.seldon.io",
        "v1",
        namespace,
        "seldondeployments",
    )
    for sdep in sdeps.get("items", []):
        sa_names = []
        for predictor in sdep.get("spec", {}).get("predictors", []):
            sa_name = predictor.get("graph", {}).get("serviceAccountName", None)
            sa_names.append(sa_name)
        output.append((sdep, sa_names))
    return output


def find_sa_related_secret(sa_name, namespace) -> Union[client.V1Secret, None]:
    """Find AWS secret related to specified SA.

    Parameters
    ----------
    sa_name: str
        Name of Service Account
    namespace:
        Name of namespace that contains the SA.

    Returns
    -------
    secret: client.V1Secret
    """
    v1 = client.CoreV1Api()
    service_account = v1.read_namespaced_service_account(sa_name, namespace)
    for s in service_account.secrets:
        secret = v1.read_namespaced_secret(s.name, namespace)
        if not all(key in secret.data for key in AWS_SA_SECRET_REQUIRED_FIELDS):
            continue
        if not all(
            key in secret.metadata.annotations
            for key in AWS_SA_SECRET_REQUIRED_ANNOTATIONS
        ):
            continue
        return secret
    return None


def new_field_for_sa_secret(secret: client.V1Secret, provider: str):
    """Get new fields that need to be added to secret.

    Parameters
    ----------
    secret: client.V1Secret
        Kubernetes secret that needs to be updated
    provider: str
        S3 provider: must be minio or aws

    Returns
    -------
    new_fields: dict
        New fields for  the secret partitioned into 'data' and 'stringData' fields
    """
    for key in AWS_SA_SECRET_REQUIRED_FIELDS:
        if key not in secret.data:
            raise ValueError(
                f"Secret '{secret.metadata.name}' does not contain '{key}' field."
            )

    use_https = secret.metadata.annotations.get(
        "machinelearning.seldon.io/s3-usehttps", None
    )
    if use_https == "0":
        protocol = "http"
    elif use_https == "1":
        protocol = "https"
    else:
        raise ValueError(
            f"Cannot determine http(s) protocol for {secret.metadata.name}."
        )

    s3_endpoint = secret.metadata.annotations.get(
        "machinelearning.seldon.io/s3-endpoint", None
    )
    if s3_endpoint is None:
        raise ValueError(f"Cannot determine S3 endpoint for {secret.metadata.name}.")

    endpoint = f"{protocol}://{s3_endpoint}"

    return {
        "data": {
            "RCLONE_CONFIG_S3_ACCESS_KEY_ID": secret.data.get("awsAccessKeyID"),
            "RCLONE_CONFIG_S3_SECRET_ACCESS_KEY": secret.data.get("awsSecretAccessKey"),
        },
        "stringData": {
            "RCLONE_CONFIG_S3_TYPE": "s3",
            "RCLONE_CONFIG_S3_PROVIDER": provider,
            "RCLONE_CONFIG_S3_ENV_AUTH": "false",
            "RCLONE_CONFIG_S3_ENDPOINT": endpoint,
        },
    }


def update_aws_sa_resources(namespaces, provider):
    """Updated AWS secrets used by Seldon Deployments via related Service Accounts in specified namespaces.

    Parameters
    ----------
    namespaces: List[str]
        List of namespaces in which will look for Seldon Deployments
    provider: str
        S3 provider: must be minio or aws
    """
    v1 = client.CoreV1Api()
    api_instance = client.CustomObjectsApi()
    for namespace in namespaces:
        print(f"Upgrading namespace {namespace}")
        for sdep, sa_names_per_predictor in get_sdeps_with_service_accounts(namespace):
            if all(sa_name is None for sa_name in sa_names_per_predictor):
                continue
            update_body = {"spec": sdep["spec"]}
            for n, sa_name in enumerate(sa_names_per_predictor):
                if sa_name is None:
                    continue
                secret = find_sa_related_secret(sa_name, namespace)
                if secret is None:
                    print(
                        f"Couldn't find secret with S3 credentials for {sa.metadata.name}"
                    )
                    continue
                new_fields = new_field_for_sa_secret(secret, "minio")
                _ = v1.patch_namespaced_secret(
                    secret.metadata.name,
                    namespace,
                    client.V1Secret(
                        data=new_fields["data"], string_data=new_fields["stringData"]
                    ),
                )
                print(f"  Upgraded secret {secret.metadata.name}")
                update_body["spec"]["predictors"][n]["graph"][
                    "envSecretRefName"
                ] = secret.metadata.name
            api_instance.patch_namespaced_custom_object(
                "machinelearning.seldon.io",
                "v1",
                namespace,
                "seldondeployments",
                sdep["metadata"]["name"],
                update_body,
            )
            print(f"  Upgrade sdep {sdep['metadata']['name']}")
```


```python
update_aws_sa_resources(namespaces=["seldon"], provider="minio")
```

## Upgrade Seldon Core to use new storage initializer


```bash
%%bash
helm upgrade seldon-core seldon-core-operator \
    --install \
    --repo https://storage.googleapis.com/seldon-charts \
    --version 1.9.1 \
    --namespace seldon-system \
    --set storageInitializer.image="seldonio/rclone-storage-initializer:1.19.0-dev" \
    --reuse-values
```


```bash
%%bash
kubectl rollout restart -n seldon-system deployments/seldon-controller-manager
kubectl rollout status -n seldon-system deployments/seldon-controller-manager
```


```python
from time import sleep

sleep(10)
```


```bash
%%bash

kubectl rollout restart deploy/$(kubectl get deploy -l seldon-deployment-id=sklearn-iris-secret -o jsonpath='{.items[0].metadata.name}')
kubectl rollout restart deploy/$(kubectl get deploy -l seldon-deployment-id=sklearn-iris-sa -o jsonpath='{.items[0].metadata.name}')

kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=sklearn-iris-secret -o jsonpath='{.items[0].metadata.name}')
kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=sklearn-iris-sa -o jsonpath='{.items[0].metadata.name}')
```


```bash
%%bash
curl -s -X POST -H 'Content-Type: application/json' \
    -d '{"data":{"ndarray":[[5.964, 4.006, 2.081, 1.031]]}}' \
    http://localhost:8003/seldon/seldon/sklearn-iris-secret/api/v1.0/predictions  | jq .
```


```bash
%%bash
curl -s -X POST -H 'Content-Type: application/json' \
    -d '{"data":{"ndarray":[[5.964, 4.006, 2.081, 1.031]]}}' \
    http://localhost:8003/seldon/seldon/sklearn-iris-sa/api/v1.0/predictions  | jq .
```

## Cleanup


```bash
%%bash
kubectl delete -f sklearn-iris-sa.yaml || echo "already removed"
kubectl delete -f sklearn-iris-secret.yaml || echo "already removed"
```
