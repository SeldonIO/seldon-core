module github.com/seldonio/seldon-core/executor

go 1.12

require (
	github.com/cloudevents/sdk-go v0.10.2
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.1.0
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/golang/protobuf v1.3.2
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.3
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/onsi/gomega v1.7.0
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.3.0
	github.com/prometheus/common v0.7.0
	github.com/seldonio/seldon-core/operator v0.0.0-20191223100430-3b372610589b
	github.com/tensorflow/tensorflow v1.14.0 // indirect
	github.com/tensorflow/tensorflow/tensorflow/go/core v0.0.0-00010101000000-000000000000
	github.com/uber/jaeger-client-go v2.21.1+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible // indirect
	golang.org/x/net v0.0.0-20190926025831-c00fd9afed17 // indirect
	golang.org/x/xerrors v0.0.0-20190717185122-a985d3407aa7
	google.golang.org/grpc v1.24.0
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.2.2
)

replace github.com/tensorflow/tensorflow/tensorflow/go/core => ./proto/tensorflow/core
