/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pipeline

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"google.golang.org/grpc/metadata"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/envoy/resources"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
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

func addRequestIdToKafkaHeadersIfMissing(headers []kafka.Header, requestId string) []kafka.Header {
	for _, kafkaHeader := range headers {
		if kafkaHeader.Key == util.RequestIdHeader { //already exists
			return headers
		}
	}
	return append(headers, kafka.Header{
		Key:   util.RequestIdHeader,
		Value: []byte(requestId),
	})
}

// We ensure the Kafka headers are lower case as http headers may have been canonical uppercased
func convertHttpHeadersToKafkaHeaders(httpHeaders http.Header) []kafka.Header {
	var kafkaHeaders []kafka.Header
	for k, vals := range httpHeaders {
		key := strings.ToLower(k)
		if strings.HasPrefix(key, resources.ExternalHeaderPrefix) {
			for _, headerValue := range vals {
				kafkaHeaders = append(kafkaHeaders, kafka.Header{Key: key, Value: []byte(headerValue)})
			}
		}
	}
	return kafkaHeaders
}

func convertKafkaHeadersToHttpHeaders(kafkaHeaders []kafka.Header) http.Header {
	httpHeaders := make(http.Header)
	for _, kafkaHeader := range kafkaHeaders {
		if strings.HasPrefix(strings.ToLower(kafkaHeader.Key), resources.ExternalHeaderPrefix) {
			if val, ok := httpHeaders[kafkaHeader.Key]; ok {
				val = append(val, string(kafkaHeader.Value))
				httpHeaders[kafkaHeader.Key] = val
			} else {
				httpHeaders[kafkaHeader.Key] = []string{string(kafkaHeader.Value)}
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
	for _, kafkaHeader := range kafkaHeaders {
		if strings.HasPrefix(strings.ToLower(kafkaHeader.Key), resources.ExternalHeaderPrefix) {
			if val, ok := grpcMetadata[kafkaHeader.Key]; ok {
				val = append(val, string(kafkaHeader.Value))
				grpcMetadata[kafkaHeader.Key] = val
			} else {
				grpcMetadata[kafkaHeader.Key] = []string{string(kafkaHeader.Value)}
			}
		}
	}
	return grpcMetadata
}
