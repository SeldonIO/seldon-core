package io.seldon.dataflow.kafka

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

fun getKafkaProperties(params: KafkaStreamsParams): KafkaProperties {
    // See https://docs.confluent.io/platform/current/streams/developer-guide/config-streams.html

    return Properties().apply {
        // TODO - add version to app ID?  (From env var.)
        this[StreamsConfig.APPLICATION_ID_CONFIG] = "seldon-dataflow-transformer"
        this[StreamsConfig.BOOTSTRAP_SERVERS_CONFIG] = params.bootstrapServers
        this[StreamsConfig.PROCESSING_GUARANTEE_CONFIG] = "at_least_once"
        this[StreamsConfig.NUM_STREAM_THREADS_CONFIG] = params.numCores * 16
        this[StreamsConfig.SECURITY_PROTOCOL_CONFIG] = "PLAINTEXT"

        // Testing
        this[StreamsConfig.REPLICATION_FACTOR_CONFIG] = 1
        this[StreamsConfig.CACHE_MAX_BYTES_BUFFERING_CONFIG] = 0
    }
}

fun KafkaProperties.withAppId(name: String): KafkaProperties {
    val properties = KafkaProperties()

    properties.putAll(this.toMap())
    val appIdPrefix = this[StreamsConfig.APPLICATION_ID_CONFIG] as String
    this[StreamsConfig.APPLICATION_ID_CONFIG] = "$appIdPrefix-$name"

    return properties
}