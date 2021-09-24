# Seldon Envoy XDS Server

## gRPC compile

Install [protoc](https://github.com/protocolbuffers/protobuf/releases).

```
protoc --version
libprotoc 3.18.0
```

Intall [go grpc plugins](https://grpc.io/docs/languages/go/quickstart/)

```
$ go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
$ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1
```

## Local Envoy Install

Install or [download](https://archive.tetratelabs.io/envoy/envoy-versions.json) envoy.

Tested with:

```
envoy --version

envoy  version: a2a1e3eed4214a38608ec223859fcfa8fb679b14/1.19.1/Clean/RELEASE/BoringSSL
```


## Smoke Test with MLServer

Install KEDA. Todo: create Ansible script. Only needed for scaling tests.

```
make install-keda
```


```
kubectl create namespace seldon-mesh
```

Install MlServer first with 3 replicas serving a hardwired Iris model.

```
make deploy-mlserver
```

When mlserver is up you need to find the 3 `endpoints` and set the ip addresses in the configmap in `k8s/scheduler/configmap.yaml`. Then deploy seldon-mesh:

```
make deploy
```

The current scheduler will use configmap to update Envoy and allow correct routing of model.

Test with curl (you will need to update ip address of svc mesh loadbalancer IP):

```
make curl-mlserver-test
```
