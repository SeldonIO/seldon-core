module github.com/seldonio/seldon-core/operator

go 1.16

require (
	github.com/Masterminds/semver v1.5.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.4.0
	github.com/gogo/protobuf v1.3.2
	github.com/google/go-cmp v0.5.5
	github.com/kedacore/keda v0.0.0-20200911122749-717aab81817f
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.14.0
	go.uber.org/zap v1.19.0
	gopkg.in/yaml.v2 v2.4.0
	istio.io/api v0.0.0-20200513175333-ae3da0d240e3
	istio.io/client-go v0.0.0-20200513180646-f8d9d8ff84e6
	k8s.io/api v0.21.3
	k8s.io/apiextensions-apiserver v0.21.3
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v12.0.0+incompatible
	knative.dev/pkg v0.0.0-20210426101439-2a0fc657a712
	sigs.k8s.io/controller-runtime v0.9.6
)

replace k8s.io/client-go => k8s.io/client-go v0.21.3

exclude github.com/go-logr/logr v1.0.0
