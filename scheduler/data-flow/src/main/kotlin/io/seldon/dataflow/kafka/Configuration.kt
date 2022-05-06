package io.seldon.dataflow.kafka

import org.apache.kafka.clients.consumer.ConsumerConfig
import org.apache.kafka.clients.producer.ProducerConfig
import org.apache.kafka.common.config.TopicConfig
import org.apache.kafka.streams.StreamsConfig
import java.util.*

data class KafkaStreamsParams(
    val bootstrapServers: String,
    val numCores: Int,
)

data class KafkaDomainParams(
    val useCleanState: Boolean,
    val joinWindowMillis: Long,
)

val KAFKA_MAX_MESSAGE_BYTES = 1_000_000_000

val kafkaTopicConfig = mapOf(
    TopicConfig.MAX_MESSAGE_BYTES_CONFIG to KAFKA_MAX_MESSAGE_BYTES.toString()
)

fun getKafkaProperties(params: KafkaStreamsParams): KafkaProperties {
    // See https://docs.confluent.io/platform/current/streams/developer-guide/config-streams.html


    return Properties().apply {
        // TODO - add version to app ID?  (From env var.)
        this[StreamsConfig.APPLICATION_ID_CONFIG] = "seldon-dataflow"
        this[StreamsConfig.BOOTSTRAP_SERVERS_CONFIG] = params.bootstrapServers
        this[StreamsConfig.PROCESSING_GUARANTEE_CONFIG] = "at_least_once"
        this[StreamsConfig.NUM_STREAM_THREADS_CONFIG] = 1
        this[StreamsConfig.SECURITY_PROTOCOL_CONFIG] = "PLAINTEXT"
        // Testing
        this[StreamsConfig.REPLICATION_FACTOR_CONFIG] = 1
        this[StreamsConfig.CACHE_MAX_BYTES_BUFFERING_CONFIG] = 0
        this[StreamsConfig.COMMIT_INTERVAL_MS_CONFIG] = 1

        this[ConsumerConfig.AUTO_OFFSET_RESET_CONFIG] = "latest"
        this[ConsumerConfig.MAX_PARTITION_FETCH_BYTES_CONFIG] = KAFKA_MAX_MESSAGE_BYTES
        this[ConsumerConfig.FETCH_MAX_BYTES_CONFIG] = KAFKA_MAX_MESSAGE_BYTES


        this[ProducerConfig.LINGER_MS_CONFIG] = 0
        this[ProducerConfig.MAX_REQUEST_SIZE_CONFIG] = KAFKA_MAX_MESSAGE_BYTES
    }
}

fun KafkaProperties.withAppId(name: String): KafkaProperties {
    val properties = KafkaProperties()

    properties.putAll(this.toMap())
    this[StreamsConfig.APPLICATION_ID_CONFIG] = "seldon-dataflow-$name"

    return properties
}
