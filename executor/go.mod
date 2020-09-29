module github.com/seldonio/seldon-core/executor

go 1.12

require (
	github.com/cloudevents/sdk-go v1.2.0
	github.com/confluentinc/confluent-kafka-go v1.4.2
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.2.0
	github.com/golang/protobuf v1.4.2
	github.com/google/uuid v1.1.2
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.1
	github.com/onsi/gomega v1.10.1
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/common v0.13.0
	github.com/seldonio/seldon-core/operator v0.0.0-00010101000000-000000000000
	github.com/tensorflow/tensorflow/tensorflow/go/core v0.0.0-00010101000000-000000000000
	github.com/uber/jaeger-client-go v2.25.0+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible // indirect
	go.uber.org/zap v1.15.0
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543
	google.golang.org/grpc v1.32.0
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.19.2
	sigs.k8s.io/controller-runtime v0.6.2
)

replace github.com/tensorflow/tensorflow/tensorflow/go/core => ./proto/tensorflow/core

replace github.com/seldonio/seldon-core/operator => ./_operator
