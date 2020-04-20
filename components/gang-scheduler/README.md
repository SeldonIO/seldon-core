
# Optimized Gang Scheduling for Seldon Core

This package encompasses an optimized module that is built to process batch loads across Seldon Core models in an optimized and robust way.

## Getting Started

Once installed, to find the available commands type:

```console
$ seldon-batch-processor --deployment-name sklearn-deployment \
    --gateway ambassador \
    --endpoint istio-ingressgateway.istio-system.svc.cluster.local:80 \
    --namespace default \
    --transport rest \
    --payload-type ndarray \
    --parallelism 10 \
    --retries 3 \
    --input-path /assets/input-data.txt \
    --output-path /assets/output-data.txt \
    --method predict \
    --debug true
```

You can also browse the [`seldon-batch` documentation](docs/sd.md).

### Docker

To make it easier to run `sd` we also offer a Docker image, which can be found
in `seldonio/seldon-batch`, which has been built to be used together with Workflow Manager frameworks like Volcano and Argo Workflows.

To use it, just type:

```console
$ docker run -it --rm seldonio/seldon-batch:latest
```

