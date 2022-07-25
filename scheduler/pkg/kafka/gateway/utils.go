package gateway

import (
	"net/http"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/resources"
	"google.golang.org/grpc/metadata"
)

func extractHeadersHttp(headers http.Header) map[string][]string {
	filteredHeaders := make(map[string][]string)
	for k, v := range headers {
		if strings.HasPrefix(k, resources.ExternalHeaderPrefix) {
			filteredHeaders[k] = v
		}
	}
	return filteredHeaders
}

func extractHeadersGrpc(headers metadata.MD, trailers metadata.MD) map[string][]string {
	filteredHeaders := make(map[string][]string)
	for k, v := range headers {
		if strings.HasPrefix(k, resources.ExternalHeaderPrefix) {
			filteredHeaders[k] = v
		}
	}
	for k, v := range trailers {
		if strings.HasPrefix(k, resources.ExternalHeaderPrefix) {
			filteredHeaders[k] = v
		}
	}
	return filteredHeaders
}

func filterKafkaInputHeaders(headers []kafka.Header) []kafka.Header {
	var filteredHeaders []kafka.Header
	for _, h := range headers {
		if strings.HasPrefix(h.Key, resources.ExternalHeaderPrefix) {
			filteredHeaders = append(filteredHeaders, h)
		}
	}
	return filteredHeaders
}
