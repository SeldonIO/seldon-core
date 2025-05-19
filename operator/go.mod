module github.com/seldonio/seldon-core/operator/v2

go 1.24.0

require (
	emperror.dev/errors v0.8.1
	github.com/banzaicloud/k8s-objectmatcher v1.8.0
	github.com/confluentinc/confluent-kafka-go/v2 v2.8.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v1.4.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0
	github.com/imdario/mergo v0.3.16
	github.com/json-iterator/go v1.1.12
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.36.2
	github.com/seldonio/seldon-core/apis/go/v2 v2.8.5
	github.com/seldonio/seldon-core/components/kafka/v2 v2.0.0-00010101000000-000000000000
	github.com/seldonio/seldon-core/components/tls/v2 v2.8.5
	github.com/spf13/cobra v1.8.1
	github.com/spf13/pflag v1.0.6
	github.com/tidwall/gjson v1.18.0
	golang.org/x/exp v0.0.0-20250128182459-e0ece0dbea4c
	google.golang.org/grpc v1.70.0
	google.golang.org/protobuf v1.36.5
	k8s.io/api v0.33.1
	k8s.io/apimachinery v0.33.1
	k8s.io/client-go v0.32.1
	k8s.io/utils v0.0.0-20241210054802-24370beab758
	knative.dev/pkg v0.0.0-20250128013458-efddeac3ec35
	sigs.k8s.io/controller-runtime v0.20.1
)

require (
	github.com/emicklei/go-restful/v3 v3.12.1 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/gnostic-models v0.6.9 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	golang.org/x/sync v0.12.0 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.4 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/evanphx/json-patch v5.9.0+incompatible // indirect
	github.com/evanphx/json-patch/v5 v5.9.11 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/go-logr/zapr v1.3.0 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nxadm/tail v1.4.11 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.20.5 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.62.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/oauth2 v0.25.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/term v0.30.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	golang.org/x/time v0.9.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.4.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250127172529-29210b9bc287 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	istio.io/pkg v0.0.0-20241216214326-d70796207df3 // indirect
	k8s.io/apiextensions-apiserver v0.32.1 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kube-openapi v0.0.0-20250318190949-c8a335a9a2ff // indirect
	sigs.k8s.io/json v0.0.0-20241014173422-cfa47c3a1cc8 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.6.0 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)

replace github.com/seldonio/seldon-core/components/tls/v2 => ../components/tls

replace github.com/seldonio/seldon-core/components/kafka/v2 => ../components/kafka

replace github.com/seldonio/seldon-core/apis/go/v2 => ../apis/go
