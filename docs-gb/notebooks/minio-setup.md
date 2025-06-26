# Install MinIO in cluster

## Helm install minio

```bash
kubectl create ns minio-system
helm repo add minio https://helm.min.io/
helm install minio minio/minio \
    --set accessKey=minioadmin \
    --set secretKey=minioadmin \
    --namespace minio-system
```

```bash
kubectl rollout status deployment -n minio-system minio
```

## port-forward Minio to localhost

In a separate terminal:

```bash
kubectl port-forward -n minio-system svc/minio 8090:9000
```

or follow instructions printed by helm

## Install MinIO CLI client tool

Install minio using `go get`:

```bash
GO111MODULE=on go get github.com/minio/mc
```

Or follow steps relevant to your platform from official [documentation](https://docs.min.io/docs/minio-client-quickstart-guide.html).

## Configure mc client to talk to your cluster

```bash
mc config host add minio-seldon http://localhost:8090 minioadmin minioadmin
``` 