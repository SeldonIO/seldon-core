module github.com/seldonio/seldon-core/executor

go 1.12

require (
	github.com/DataDog/datadog-go v4.0.0+incompatible // indirect
	github.com/cloudevents/sdk-go v1.2.0
	github.com/confluentinc/confluent-kafka-go v1.4.2
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.1.0
	github.com/golang/protobuf v1.4.2
	github.com/google/uuid v1.1.2
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.1
	github.com/onsi/gomega v1.10.1
	github.com/opentracing/opentracing-go v1.2.0
	github.com/philhofer/fwd v1.0.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/common v0.13.0
	github.com/seldonio/seldon-core/operator v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.5.1 // indirect
	github.com/tensorflow/tensorflow/tensorflow/go/core v0.0.0-00010101000000-000000000000
	github.com/uber/jaeger-client-go v2.25.0+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible // indirect
	go.uber.org/zap v1.15.0
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a // indirect
	golang.org/x/net v0.0.0-20200822124328-c89045814202 // indirect
	golang.org/x/sys v0.0.0-20200826173525-f9321e4c35a6 // indirect
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543
	google.golang.org/grpc v1.31.0
	gopkg.in/DataDog/dd-trace-go.v1 v1.26.0
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.17.9
	k8s.io/apimachinery v0.17.9
	sigs.k8s.io/controller-runtime v0.5.8
)

replace github.com/tensorflow/tensorflow/tensorflow/go/core => ./proto/tensorflow/core

replace github.com/seldonio/seldon-core/operator => ./_operator
