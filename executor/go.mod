module github.com/seldonio/seldon-core/executor

go 1.12

require (
	github.com/cloudevents/sdk-go v1.2.0
	github.com/fullstorydev/grpcurl v1.5.1 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.1.0
	github.com/golang/protobuf v1.4.2
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.4
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/onsi/gomega v1.8.1
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.3.0
	github.com/prometheus/common v0.10.0
	github.com/seldonio/seldon-core/operator v0.0.0-20200401123312-d4c435ea5217
	github.com/soheilhy/cmux v0.1.4
	github.com/tensorflow/tensorflow v1.14.0 // indirect
	github.com/tensorflow/tensorflow/tensorflow/go/core v0.0.0-00010101000000-000000000000
	github.com/uber/jaeger-client-go v2.21.1+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible // indirect
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543
	google.golang.org/grpc v1.30.0
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
	sigs.k8s.io/controller-runtime v0.5.0
	sigs.k8s.io/controller-tools v0.2.0 // indirect
	sigs.k8s.io/testing_frameworks v0.1.1 // indirect
)

replace github.com/tensorflow/tensorflow/tensorflow/go/core => ./proto/tensorflow/core
