module github.com/seldonio/seldon-core/operator

go 1.12

require (
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20190717042225-c3de453c63f4 // indirect
	github.com/go-logr/logr v0.1.0
	github.com/google/go-cmp v0.3.1
	github.com/knative/pkg v0.0.0-20190823221514-39a29cf1bf26
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/prometheus/common v0.0.0-20180801064454-c7de2306084e
	github.com/seldonio/seldon-operator v0.4.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6 // indirect
	gopkg.in/yaml.v2 v2.2.2
	k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	knative.dev/pkg v0.0.0-20190823221514-39a29cf1bf26
	sigs.k8s.io/controller-runtime v0.2.0
	sigs.k8s.io/controller-tools v0.2.0 // indirect
	sigs.k8s.io/kind v0.5.1 // indirect
	sigs.k8s.io/yaml v1.1.0
)
