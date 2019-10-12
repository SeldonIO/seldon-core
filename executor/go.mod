module github.com/seldonio/seldon-core/executor

go 1.12

require (
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20190924025748-f65c72e2690d // indirect
	github.com/go-logr/logr v0.1.0
	github.com/gogo/protobuf v1.1.1
	github.com/golang/protobuf v1.3.2
	github.com/gorilla/mux v1.7.3
	github.com/onsi/gomega v1.7.0
	github.com/pkg/errors v0.8.1
	github.com/prometheus/common v0.0.0-20180801064454-c7de2306084e
	github.com/sirupsen/logrus v1.4.2 // indirect
	github.com/tensorflow/tensorflow v1.14.0
	github.com/tensorflow/tensorflow/tensorflow/go/core v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.0.0-20190926025831-c00fd9afed17 // indirect
	golang.org/x/xerrors v0.0.0-20190717185122-a985d3407aa7
	google.golang.org/grpc v1.24.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6 // indirect
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.2.2
)

replace github.com/tensorflow/tensorflow/tensorflow/go/core => ./proto/tensorflow/core
