module github.com/seldonio/seldon-core/operator/v2

go 1.21

require (
	emperror.dev/errors v0.8.1
	github.com/banzaicloud/k8s-objectmatcher v1.8.0
	github.com/confluentinc/confluent-kafka-go/v2 v2.5.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v1.4.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0
	github.com/imdario/mergo v0.3.16
	github.com/json-iterator/go v1.1.12
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.33.1
	github.com/seldonio/seldon-core/apis/go/v2 v2.0.0-00010101000000-000000000000
	github.com/seldonio/seldon-core/components/tls/v2 v2.0.0-00010101000000-000000000000
	github.com/spf13/cobra v1.8.0
	github.com/spf13/pflag v1.0.5
	github.com/tidwall/gjson v1.17.1
	google.golang.org/grpc v1.65.0
	google.golang.org/protobuf v1.34.1
	k8s.io/api v0.29.2
	k8s.io/apimachinery v0.29.2
	k8s.io/client-go v0.29.2
	k8s.io/utils v0.0.0-20230726121419-3b25d923346b
	knative.dev/pkg v0.0.0-20211203062937-d37811b71d6a
	sigs.k8s.io/controller-runtime v0.17.4
)

require (
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/matttproud/golang_protobuf_extensions/v2 v2.0.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	golang.org/x/exp v0.0.0-20240112132812-db7319d0e0e3 // indirect
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.3 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/evanphx/json-patch v4.12.0+incompatible // indirect
	github.com/evanphx/json-patch/v5 v5.8.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-logr/zapr v1.3.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.18.0 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.45.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/oauth2 v0.20.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
	golang.org/x/term v0.20.0 // indirect
	golang.org/x/text v0.15.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.4.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240528184218-531527333157 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	istio.io/pkg v0.0.0-20220805133948-6c7080d8dec5 // indirect
	k8s.io/apiextensions-apiserver v0.29.2 // indirect
	k8s.io/component-base v0.29.2 // indirect
	k8s.io/klog/v2 v2.110.1 // indirect
	k8s.io/kube-openapi v0.0.0-20231010175941-2dd684a91f00 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)

replace github.com/seldonio/seldon-core/components/tls/v2 => ../components/tls

replace github.com/seldonio/seldon-core/apis/go/v2 => ../apis/go
