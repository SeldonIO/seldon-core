module github.com/seldonio/seldon-core/scheduler

go 1.16

require (
	github.com/cenkalti/backoff/v4 v4.1.1
	github.com/envoyproxy/go-control-plane v0.9.10-0.20210910171841-453346fa5903
	github.com/go-playground/validator/v10 v10.9.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/jarcoal/httpmock v1.0.8
	github.com/onsi/gomega v1.16.0
	github.com/sirupsen/logrus v1.8.1
	google.golang.org/grpc v1.38.0
	google.golang.org/protobuf v1.26.0
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	sigs.k8s.io/controller-runtime v0.10.3
)
