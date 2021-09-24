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