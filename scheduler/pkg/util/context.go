/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package util

import (
	"context"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/v2/kafka/splunkkafka"
	"go.opentelemetry.io/otel"
)

// Extract tracing context from Kafka message
func CreateContextFromKafkaMsg(msg *kafka.Message) (context.Context, context.CancelFunc) {
	ctx := context.Background()
	carrierIn := splunkkafka.NewMessageCarrier(msg)
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrierIn)
	return context.WithTimeout(ctx, inferTimeoutDefault)
}
