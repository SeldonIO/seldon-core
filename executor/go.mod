module github.com/seldonio/seldon-core/executor

go 1.12

require (
	github.com/DataDog/datadog-go v4.0.0+incompatible // indirect
	github.com/bradfitz/gomemcache v0.0.0-20190913173617-a41fca850d0b // indirect
	github.com/cloudevents/sdk-go v1.2.0
	github.com/confluentinc/confluent-kafka-go v1.4.2
	github.com/garyburd/redigo v1.6.2 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/gin-gonic/gin v1.6.3 // indirect
	github.com/go-chi/chi v4.1.2+incompatible // indirect
	github.com/go-logr/logr v0.1.0
	github.com/go-redis/redis v6.15.9+incompatible // indirect
	github.com/gocql/gocql v0.0.0-20200815110948-5378c8f664e9 // indirect
	github.com/golang/protobuf v1.4.2
	github.com/gomodule/redigo v1.8.2 // indirect
	github.com/google/uuid v1.1.2
	github.com/gorilla/mux v1.8.0
	github.com/graph-gophers/graphql-go v0.0.0-20200819123640-3b5ddcd884ae // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.1
	github.com/hashicorp/vault/api v1.0.4 // indirect
	github.com/jinzhu/gorm v1.9.16 // indirect
	github.com/jmoiron/sqlx v1.2.0 // indirect
	github.com/labstack/echo v3.3.10+incompatible // indirect
	github.com/labstack/echo/v4 v4.1.17 // indirect
	github.com/onsi/gomega v1.10.1
	github.com/opentracing/opentracing-go v1.2.0
	github.com/philhofer/fwd v1.0.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/common v0.13.0
	github.com/seldonio/seldon-core/operator v0.0.0-00010101000000-000000000000
	github.com/soheilhy/cmux v0.1.4
	github.com/syndtr/goleveldb v1.0.0 // indirect
	github.com/tensorflow/tensorflow/tensorflow/go/core v0.0.0-00010101000000-000000000000
	github.com/tidwall/buntdb v1.1.2 // indirect
	github.com/twitchtv/twirp v5.12.1+incompatible // indirect
	github.com/uber/jaeger-client-go v2.25.0+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible // indirect
	github.com/zenazn/goji v1.0.1 // indirect
	go.uber.org/zap v1.15.0
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543
	google.golang.org/grpc v1.31.0
	gopkg.in/DataDog/dd-trace-go.v1 v1.26.0
	gopkg.in/confluentinc/confluent-kafka-go.v1 v1.4.2 // indirect
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.17.9
	k8s.io/apimachinery v0.17.9
	sigs.k8s.io/controller-runtime v0.5.8
)

replace github.com/tensorflow/tensorflow/tensorflow/go/core => ./proto/tensorflow/core

replace github.com/seldonio/seldon-core/operator => ./_operator
