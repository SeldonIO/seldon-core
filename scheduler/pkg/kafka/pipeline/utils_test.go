package pipeline

import (
	"net/http"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"google.golang.org/grpc/metadata"

	. "github.com/onsi/gomega"
)

func TestCreateResourceNameFromHeader(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                 string
		header               string
		expectedResourceName string
		isModel              bool
		error                bool
	}
	tests := []test{
		{
			name:                 "model no suffix",
			header:               "foo",
			expectedResourceName: "foo",
			isModel:              true,
		},
		{
			name:                 "pipeline",
			header:               "foo.pipeline",
			expectedResourceName: "foo",
			isModel:              false,
		},
		{
			name:                 "model with suffix",
			header:               "foo.model",
			expectedResourceName: "foo",
			isModel:              true,
		},
		{
			name:   "model with too many parts",
			header: "foo.bar.model",
			error:  true,
		},
		{
			name:   "bad suffix",
			header: "foo.bar",
			error:  true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resouceName, isModel, err := createResourceNameFromHeader(test.header)
			if test.error {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(resouceName).To(Equal(test.expectedResourceName))
				g.Expect(isModel).To(Equal(test.isModel))
			}
		})
	}
}

func TestConvertHttpHeadersToKafkaHeaders(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                 string
		httpHeaders          http.Header
		expectedKafkaHeaders map[string]kafka.Header
	}
	tests := []test{
		{
			name: "example http headers to kafka headers",
			httpHeaders: http.Header{
				"Content-Type": []string{"json"},
				"X-foo":        []string{"bar"},
				"X-foo2":       []string{"bar2"},
			},
			expectedKafkaHeaders: map[string]kafka.Header{
				"X-foo":  {Key: "x-foo", Value: []byte("bar")},
				"X-foo2": {Key: "x-foo2", Value: []byte("bar2")},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			kafkaHeaders := convertHttpHeadersToKafkaHeaders(test.httpHeaders)
			for _, kafkaHeader := range kafkaHeaders {
				v, ok := test.expectedKafkaHeaders[kafkaHeader.Key]
				g.Expect(ok).To(BeTrue())
				g.Expect(kafkaHeader.Value).To(Equal(v.Value))
			}
		})
	}
}

func TestConvertKafkaHeadersToHttpHeaders(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                string
		kafkaHeaders        []kafka.Header
		expectedHttpHeaders http.Header
	}
	tests := []test{
		{
			name: "example kafka headers to http headers",
			kafkaHeaders: []kafka.Header{
				{Key: "X-foo", Value: []byte("bar")},
				{Key: "Content-Type", Value: []byte("json")},
			},
			expectedHttpHeaders: http.Header{
				"X-foo": []string{"bar"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			httpHeaders := convertKafkaHeadersToHttpHeaders(test.kafkaHeaders)
			for k, v := range httpHeaders {
				vExpected, ok := test.expectedHttpHeaders[k]
				g.Expect(ok).To(BeTrue())
				g.Expect(v).To(Equal(vExpected))
			}
		})
	}
}

func TestConvertGrpcMetadataToKafkaHeaders(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                 string
		meta                 metadata.MD
		expectedKafkaHeaders map[string]kafka.Header
	}
	tests := []test{
		{
			name: "example meta to kafka headers",
			meta: metadata.MD{
				"X-foo":        []string{"bar"},
				"Content-Type": []string{"json"},
			},
			expectedKafkaHeaders: map[string]kafka.Header{
				"X-foo": {Key: "X-foo", Value: []byte("bar")},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			kafkaHeaders := convertGrpcMetadataToKafkaHeaders(test.meta)
			for _, kafkaHeader := range kafkaHeaders {
				v, ok := test.expectedKafkaHeaders[kafkaHeader.Key]
				g.Expect(ok).To(BeTrue())
				g.Expect(kafkaHeader.Value).To(Equal(v.Value))
			}
		})
	}
}

func TestConvertKafkaHeadersToGrpcMetadata(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name         string
		kafkaHeaders []kafka.Header
		expectedMeta metadata.MD
	}
	tests := []test{
		{
			name: "example kafka headers to grpc headers",
			kafkaHeaders: []kafka.Header{
				{Key: "X-foo", Value: []byte("bar")},
				{Key: "Content-Type", Value: []byte("json")},
			},
			expectedMeta: metadata.MD{
				"X-foo": []string{"bar"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			httpHeaders := convertKafkaHeadersToHttpHeaders(test.kafkaHeaders)
			for k, v := range httpHeaders {
				vExpected, ok := test.expectedMeta[k]
				g.Expect(ok).To(BeTrue())
				g.Expect(v).To(Equal(vExpected))
			}
		})
	}
}
