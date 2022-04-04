module github.com/seldonio/seldon-core/scheduler

go 1.16

require (
	github.com/cenkalti/backoff/v4 v4.1.1
	github.com/confluentinc/confluent-kafka-go v1.8.2
	github.com/envoyproxy/go-control-plane v0.10.1
	github.com/fsnotify/fsnotify v1.5.1
	github.com/go-playground/validator/v10 v10.9.0
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/jarcoal/httpmock v1.0.8
	github.com/mitchellh/copystructure v1.2.0
	github.com/mustafaturan/bus/v3 v3.0.3
	github.com/onsi/gomega v1.16.0
	github.com/orcaman/concurrent-map v1.0.0 // indirect
	github.com/otiai10/copy v1.7.0
	github.com/prometheus/client_golang v1.12.1
	github.com/rs/xid v1.3.0
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	google.golang.org/grpc v1.42.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	knative.dev/pkg v0.0.0-20211207151905-681fbddaeb50
	sigs.k8s.io/controller-runtime v0.10.3
)
