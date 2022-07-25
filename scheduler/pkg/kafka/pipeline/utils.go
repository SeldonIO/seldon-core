package pipeline

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"google.golang.org/grpc/metadata"

	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/resources"
)

func createResourceNameFromHeader(header string) (string, bool, error) {
	parts := strings.Split(header, ".")
	switch len(parts) {
	case 1:
		if len(parts[0]) > 0 {
			return header, true, nil
		}
	case 2:
		switch parts[1] {
		case resources.SeldonPipelineHeaderSuffix:
			return parts[0], false, nil
		case resources.SeldonModelHeaderSuffix:
			return parts[0], true, nil
		}
	}
	return "", false, fmt.Errorf(
		"Bad or missing header %s %s", resources.SeldonModelHeader, header)
}

func convertHttpHeadersToKafkaHeaders(httpHeaders http.Header) []kafka.Header {
	var kafkaHeaders []kafka.Header
	for k, vals := range httpHeaders {
		if strings.HasPrefix(strings.ToLower(k), resources.ExternalHeaderPrefix) {
			for _, headerValue := range vals {
				kafkaHeaders = append(kafkaHeaders, kafka.Header{Key: k, Value: []byte(headerValue)})
			}
		}
	}
	return kafkaHeaders
}

func convertKafkaHeadersToHttpHeaders(kafkaHeaders []kafka.Header) http.Header {
	httpHeaders := make(http.Header)
	found := make(map[string]bool)
	for _, kafkaHeader := range kafkaHeaders {
		if strings.HasPrefix(strings.ToLower(kafkaHeader.Key), resources.ExternalHeaderPrefix) {
			if !found[kafkaHeader.Key] {
				if val, ok := httpHeaders[kafkaHeader.Key]; ok {
					val = append(val, string(kafkaHeader.Value))
					httpHeaders[kafkaHeader.Key] = val
				} else {
					httpHeaders[kafkaHeader.Key] = []string{string(kafkaHeader.Value)}
				}
				found[kafkaHeader.Key] = true
			}
		}
	}
	return httpHeaders
}

func convertGrpcMetadataToKafkaHeaders(grpcMetadata metadata.MD) []kafka.Header {
	var kafkaHeaders []kafka.Header
	for k, vals := range grpcMetadata {
		if strings.HasPrefix(strings.ToLower(k), resources.ExternalHeaderPrefix) {
			for _, headerValue := range vals {
				kafkaHeaders = append(kafkaHeaders, kafka.Header{Key: k, Value: []byte(headerValue)})
			}
		}
	}
	return kafkaHeaders
}

func convertKafkaHeadersToGrpcMetadata(kafkaHeaders []kafka.Header) metadata.MD {
	grpcMetadata := make(metadata.MD)
	found := make(map[string]bool)
	for _, kafkaHeader := range kafkaHeaders {
		if strings.HasPrefix(strings.ToLower(kafkaHeader.Key), resources.ExternalHeaderPrefix) {
			if !found[kafkaHeader.Key] {
				if val, ok := grpcMetadata[kafkaHeader.Key]; ok {
					val = append(val, string(kafkaHeader.Value))
					grpcMetadata[kafkaHeader.Key] = val
				} else {
					grpcMetadata[kafkaHeader.Key] = []string{string(kafkaHeader.Value)}
				}
				found[kafkaHeader.Key] = true
			}
		}
	}
	return grpcMetadata
}
