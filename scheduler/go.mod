module github.com/seldonio/seldon-core/scheduler

go 1.16

require (
	github.com/OneOfOne/xxhash v1.2.8 // indirect
	github.com/cenkalti/backoff/v4 v4.1.2
	github.com/confluentinc/confluent-kafka-go v1.8.2
	github.com/dgraph-io/badger/v3 v3.2103.2
	github.com/envoyproxy/go-control-plane v0.10.2-0.20220325020618-49ff273808a1
	github.com/fsnotify/fsnotify v1.5.1
	github.com/go-logr/zapr v1.2.3 // indirect
	github.com/go-playground/validator/v10 v10.9.0
	github.com/golang/protobuf v1.5.2
	github.com/google/go-cmp v0.5.7
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/jarcoal/httpmock v1.0.8
	github.com/mitchellh/copystructure v1.2.0
	github.com/mustafaturan/bus/v3 v3.0.3
	github.com/onsi/gomega v1.17.0
	github.com/orcaman/concurrent-map v1.0.0
	github.com/otiai10/copy v1.7.0
	github.com/prometheus/client_golang v1.12.1
	github.com/rs/xid v1.3.0
	github.com/serialx/hashring v0.0.0-20200727003509-22c0c7ab6b1b
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka v0.8.0
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.1
	go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux v0.31.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.31.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.31.0
	go.opentelemetry.io/otel v1.6.3
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.6.3
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.6.3
	go.opentelemetry.io/otel/sdk v1.6.3
	go.opentelemetry.io/otel/trace v1.6.3
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/zap v1.21.0 // indirect
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 // indirect
	golang.org/x/net v0.0.0-20220421235706-1d1ef9303861 // indirect
	golang.org/x/sys v0.0.0-20220422013727-9388b58f7150 // indirect
	golang.org/x/time v0.0.0-20211116232009-f0f3c7e86c11 // indirect
	google.golang.org/genproto v0.0.0-20220422154200-b37d22cd5731 // indirect
	google.golang.org/grpc v1.46.0
	google.golang.org/protobuf v1.28.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.23.5
	k8s.io/apimachinery v0.23.5
	k8s.io/client-go v0.23.5
	knative.dev/pkg v0.0.0-20211207151905-681fbddaeb50
	sigs.k8s.io/controller-runtime v0.11.2
)
