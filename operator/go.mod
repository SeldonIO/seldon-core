module github.com/seldonio/seldon-core/operatorv2

go 1.16

require (
	github.com/go-logr/logr v0.4.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/imdario/mergo v0.3.12
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.16.0
	google.golang.org/grpc v1.42.0
	google.golang.org/protobuf v1.27.1
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
	knative.dev/pkg v0.0.0-20211203062937-d37811b71d6a
	sigs.k8s.io/controller-runtime v0.10.0
)
