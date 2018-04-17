# H2O MOJO Model
An example H2O model

## Dependencies

 * [H2O python library](http://docs.h2o.ai/h2o/latest-stable/h2o-docs/downloading.html#install-in-python)

## Train locally

```bash
make train
```

## Wrap using [s2i](https://github.com/openshift/source-to-image#installation).

```bash
s2i build . seldonio/seldon-core-s2i-java-build h2o-test:0.1 --runtime-image seldonio/seldon-core-s2i-java-runtime
```

## Local Docker Smoke Test

Run under docker.

```bash
docker run --rm -p 5000:5000 h2o-test:0.1
```

Ensure test grpc modules compiled.

```bash
pushd ../../../wrappers/testing ; make build_protos ; popd
```

Send a data request using the wrapper tester.

```bash
python ../../../wrappers/testing/tester.py contract.json 0.0.0.0 5000 -p
```

## Minikube test

```bash
minikube start --memory 4096
```

[Install seldon core](/readme.md#install)

Connect to Minikube Docker daemon

```bash
eval $(minikube docker-env)
```

Build image using minikube docker daemon.

```bash
s2i build . seldonio/seldon-core-s2i-java-build h2o-test:0.1 --runtime-image seldonio/seldon-core-s2i-java-runtime
```

Launch deployment

```bash
kubectl create -f h2o_deployment.json
```

Port forward API server

```bash
kubectl port-forward $(kubectl get pods -n seldon -l app=seldon-apiserver-container-app -o jsonpath='{.items[0].metadata.name}') -n seldon 8080:8080
```

Ensure tester gRPC modules compiled. 

```bash
pushd ../../../util/api_tester ; make build_protos ; popd
```

Send test request
```bash
python ../../../util/api_tester/api-tester.py contract.json 0.0.0.0 8080 --oauth-key oauth-key --oauth-secret oauth-secret -p
```


