module github.com/seldonio/seldon-core/operator

go 1.14

require (
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.2.0
	github.com/gogo/protobuf v1.3.1
	github.com/google/go-cmp v0.5.1
	github.com/kedacore/keda v0.0.0-20200911122749-717aab81817f
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	go.uber.org/zap v1.15.0
	gopkg.in/yaml.v2 v2.3.0
	istio.io/api v0.0.0-20200513175333-ae3da0d240e3
	istio.io/client-go v0.0.0-20200513180646-f8d9d8ff84e6
	k8s.io/api v0.19.3
	k8s.io/apiextensions-apiserver v0.19.3
	k8s.io/apimachinery v0.19.3
	k8s.io/client-go v12.0.0+incompatible
	knative.dev/pkg v0.0.0-20200911145400-2d4efecc6bc1
	sigs.k8s.io/controller-runtime v0.6.3
)

replace k8s.io/client-go => k8s.io/client-go v0.18.8
