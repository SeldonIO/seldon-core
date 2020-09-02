module github.com/seldonio/seldon-core/executor

go 1.12

require (
	github.com/cloudevents/sdk-go v0.10.2
	github.com/cloudevents/sdk-go v0.10.2
	github.com/cloudevents/sdk-go v1.2.0
	github.com/confluentinc/confluent-kafka-go v1.4.2
	github.com/fullstorydev/grpcurl v1.5.1 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.1.0
	github.com/golang/protobuf v1.3.2
	github.com/golang/protobuf v1.3.5
	github.com/golang/protobuf v1.4.2
	github.com/google/uuid v1.1.1
	github.com/google/uuid v1.1.1
	github.com/google/uuid v1.1.2
	github.com/gorilla/mux v1.7.3
	github.com/gorilla/mux v1.7.3
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.1
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/onsi/gomega v1.10.1
	github.com/onsi/gomega v1.8.1
	github.com/onsi/gomega v1.8.1
	github.com/opentracing/opentracing-go v1.1.0
	github.com/opentracing/opentracing-go v1.1.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.8.1
	github.com/pkg/errors v0.8.1
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.3.0
	github.com/prometheus/client_golang v1.3.0
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/common v0.13.0
	github.com/prometheus/common v0.7.0
	github.com/prometheus/common v0.7.0
	github.com/seldonio/seldon-core/operator v0.0.0-00010101000000-000000000000
	github.com/seldonio/seldon-core/operator v0.0.0-20200401123312-d4c435ea5217
	github.com/seldonio/seldon-core/operator v0.0.0-20200401123312-d4c435ea5217
	github.com/soheilhy/cmux v0.1.4
	github.com/tensorflow/tensorflow v1.14.0 // indirect
	github.com/tensorflow/tensorflow v1.14.0 // indirect
	github.com/tensorflow/tensorflow/tensorflow/go/core v0.0.0-00010101000000-000000000000
	github.com/uber/jaeger-client-go v2.25.0+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible // indirect
	golang.org/x/xerrors v0.0.0-20190717185122-a985d3407aa7
	golang.org/x/xerrors v0.0.0-20190717185122-a985d3407aa7
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543
	google.golang.org/grpc v1.27.0
	google.golang.org/grpc v1.28.0
	google.golang.org/grpc v1.31.0
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.17.9
	k8s.io/apimachinery v0.17.9
	sigs.k8s.io/controller-runtime v0.5.8
)

replace github.com/tensorflow/tensorflow/tensorflow/go/core => ./proto/tensorflow/core

replace github.com/seldonio/seldon-core/operator => ./_operator
