/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow.kafka

import io.klogging.noCoLogger
import io.seldon.dataflow.kafka.security.SaslConfig
import io.seldon.dataflow.mtls.CertificateConfig
import io.seldon.dataflow.mtls.K8sCertSecretsProvider
import io.seldon.dataflow.mtls.Provider
import io.seldon.dataflow.sasl.SaslOauthProvider
import io.seldon.dataflow.sasl.SaslPasswordProvider
import org.apache.kafka.clients.consumer.ConsumerConfig
import org.apache.kafka.clients.producer.ProducerConfig
import org.apache.kafka.common.config.SaslConfigs
import org.apache.kafka.common.config.SslConfigs
import org.apache.kafka.common.config.TopicConfig
import org.apache.kafka.common.security.auth.SecurityProtocol
import org.apache.kafka.streams.StreamsConfig
import org.apache.kafka.streams.errors.DeserializationExceptionHandler
import org.apache.kafka.streams.errors.ProductionExceptionHandler
import org.apache.kafka.streams.errors.StreamsUncaughtExceptionHandler
import java.util.*

const val KAFKA_UNCAUGHT_EXCEPTION_HANDLER_CLASS_CONFIG = "default.processing.exception.handler"

data class KafkaStreamsParams(
    val bootstrapServers: String,
    val numPartitions: Int,
    val replicationFactor: Int,
    val maxMessageSizeBytes: Int,
    val security: KafkaSecurityParams,
)

data class KafkaSecurityParams(
    val securityProtocol: SecurityProtocol,
    val certConfig: CertificateConfig,
    val saslConfig: SaslConfig,
)

data class KafkaDomainParams(
    val useCleanState: Boolean,
    val joinWindowMillis: Long,
)

data class TopicWaitRetryParams(
    val createTimeoutMillis: Int, // int required by the underlying kafka-streams library
    val describeTimeoutMillis: Long,
    val describeRetries: Int,
    val describeRetryDelayMillis: Long
)

val kafkaTopicConfig = { maxMessageSizeBytes: Int ->
    mapOf(
        TopicConfig.MAX_MESSAGE_BYTES_CONFIG to maxMessageSizeBytes.toString(),
    )
}

private val logger = noCoLogger(Pipeline::class)

fun getKafkaAdminProperties(params: KafkaStreamsParams): KafkaAdminProperties {
    return getSecurityProperties(params).apply {
        this[StreamsConfig.BOOTSTRAP_SERVERS_CONFIG] = params.bootstrapServers
    }
}

private fun getSecurityProperties(params: KafkaStreamsParams): Properties {
    val authProperties = when (params.security.securityProtocol) {
        SecurityProtocol.SSL -> getSslProperties(params)
        SecurityProtocol.SASL_SSL -> getSaslProperties(params)
        else -> Properties() // No authentication, so nothing to configure
    }

    return authProperties.apply {
        this[StreamsConfig.SECURITY_PROTOCOL_CONFIG] = params.security.securityProtocol.toString()
    }
}

private fun getSslProperties(params: KafkaStreamsParams): Properties {
    val certConfig = params.security.certConfig

    if (certConfig.brokerSecret != "" || certConfig.clientSecret != "") {
        K8sCertSecretsProvider.downloadCertsFromSecrets(certConfig)
    }

    return Properties().apply {
        val trustStoreConfig = Provider.trustStoreFromCertificates(certConfig)
        this[SslConfigs.SSL_TRUSTSTORE_LOCATION_CONFIG] = trustStoreConfig.trustStoreLocation
        this[SslConfigs.SSL_TRUSTSTORE_PASSWORD_CONFIG] = trustStoreConfig.trustStorePassword

        if (certConfig.clientSecret.isNotEmpty()) {
            val keyStoreConfig = Provider.keyStoreFromCertificates(certConfig)
            this[SslConfigs.SSL_KEYSTORE_LOCATION_CONFIG] = keyStoreConfig.keyStoreLocation
            this[SslConfigs.SSL_KEYSTORE_PASSWORD_CONFIG] = keyStoreConfig.keyStorePassword
            this[SslConfigs.SSL_KEY_PASSWORD_CONFIG] = keyStoreConfig.keyStorePassword
        }

        this[SslConfigs.SSL_ENDPOINT_IDENTIFICATION_ALGORITHM_CONFIG] = certConfig.endpointIdentificationAlgorithm
    }
}

private fun getSaslProperties(params: KafkaStreamsParams): Properties {
    return Properties().apply {
        if (params.security.certConfig.brokerSecret != "") {
            K8sCertSecretsProvider.downloadCertsFromSecrets(params.security.certConfig)
        }

        val trustStoreConfig = Provider.trustStoreFromCertificates(params.security.certConfig)
        this[SslConfigs.SSL_TRUSTSTORE_LOCATION_CONFIG] = trustStoreConfig.trustStoreLocation
        this[SslConfigs.SSL_TRUSTSTORE_PASSWORD_CONFIG] = trustStoreConfig.trustStorePassword
        this[SslConfigs.SSL_ENDPOINT_IDENTIFICATION_ALGORITHM_CONFIG] =
            params.security.certConfig.endpointIdentificationAlgorithm

        this[SaslConfigs.SASL_MECHANISM] = params.security.saslConfig.mechanism.toString()

        when (params.security.saslConfig) {
            is SaslConfig.Password -> {
                val module = when (params.security.saslConfig) {
                    is SaslConfig.Password.Plain -> "org.apache.kafka.common.security.plain.PlainLoginModule required"
                    is SaslConfig.Password.Scram256,
                    is SaslConfig.Password.Scram512 -> "org.apache.kafka.common.security.scram.ScramLoginModule required"
                }
                val password = SaslPasswordProvider.default.getPassword(params.security.saslConfig)

                this[SaslConfigs.SASL_JAAS_CONFIG] = module +
                        """ username="${params.security.saslConfig.username}"""" +
                        """ password="$password";"""
            }
            is SaslConfig.Oauth -> {
                val oauthConfig = SaslOauthProvider.default.getOauthConfig(params.security.saslConfig)

                val jaasConfig = buildString {
                    append("org.apache.kafka.common.security.oauthbearer.OAuthBearerLoginModule required")
                    append(""" clientId="${oauthConfig.clientId}"""")
                    append(""" clientSecret="${oauthConfig.clientSecret}"""")
                    oauthConfig.scope?.let {
                        append(""" scope="$it"""")
                    }
                    oauthConfig.extensions?.let { extensions ->
                        extensions.forEach {
                            append(""" $it""")
                        }
                    }
                    append(";")
                }

                this[SaslConfigs.SASL_JAAS_CONFIG] = jaasConfig
                this[SaslConfigs.SASL_OAUTHBEARER_TOKEN_ENDPOINT_URL] = oauthConfig.tokenUrl
                this[SaslConfigs.SASL_LOGIN_CALLBACK_HANDLER_CLASS] =
                    "org.apache.kafka.common.security.oauthbearer.OAuthBearerLoginCallbackHandler"
            }
        }
    }
}

fun getKafkaProperties(params: KafkaStreamsParams): KafkaProperties {
    // See https://docs.confluent.io/platform/current/streams/developer-guide/config-streams.html

    return getSecurityProperties(params).apply {
        // TODO - add version to app ID?  (From env var.)
        this[StreamsConfig.APPLICATION_ID_CONFIG] = "seldon-dataflow"
        this[StreamsConfig.BOOTSTRAP_SERVERS_CONFIG] = params.bootstrapServers
        this[StreamsConfig.PROCESSING_GUARANTEE_CONFIG] = StreamsConfig.AT_LEAST_ONCE
        this[StreamsConfig.NUM_STREAM_THREADS_CONFIG] = 1
        this[StreamsConfig.SEND_BUFFER_CONFIG] = params.maxMessageSizeBytes
        this[StreamsConfig.RECEIVE_BUFFER_CONFIG] = params.maxMessageSizeBytes
        // tell Kafka Streams to optimize the topology
        this[StreamsConfig.TOPOLOGY_OPTIMIZATION_CONFIG] = StreamsConfig.OPTIMIZE

        // Testing
        this[StreamsConfig.REPLICATION_FACTOR_CONFIG] = params.replicationFactor
        this[StreamsConfig.COMMIT_INTERVAL_MS_CONFIG] = 10_000

        this[ConsumerConfig.AUTO_OFFSET_RESET_CONFIG] = "latest"
        this[ConsumerConfig.MAX_PARTITION_FETCH_BYTES_CONFIG] = params.maxMessageSizeBytes
        this[ConsumerConfig.FETCH_MAX_BYTES_CONFIG] = params.maxMessageSizeBytes
        this[ConsumerConfig.SEND_BUFFER_CONFIG] = params.maxMessageSizeBytes
        this[ConsumerConfig.RECEIVE_BUFFER_CONFIG] = params.maxMessageSizeBytes
        this[ConsumerConfig.SESSION_TIMEOUT_MS_CONFIG] = 60_000

        this[ProducerConfig.LINGER_MS_CONFIG] = 0
        this[ProducerConfig.MAX_REQUEST_SIZE_CONFIG] = params.maxMessageSizeBytes
        this[ProducerConfig.BUFFER_MEMORY_CONFIG] = params.maxMessageSizeBytes
    }
}

fun KafkaProperties.withAppId(namespace: String, consumerGroupIdPrefix: String, name: String): KafkaProperties {
    val properties = KafkaProperties()
    properties.putAll(this.toMap())

    val appId = StringBuilder()
    if (consumerGroupIdPrefix.isNotEmpty()) {
        appId.append("$consumerGroupIdPrefix-")
    }
    if (namespace.isNotEmpty()) {
        appId.append("$namespace-")
    }
    appId.append("seldon-dataflow-$name")
    properties[StreamsConfig.APPLICATION_ID_CONFIG] = appId.toString()

    // TODO add k8s host name to ensure static membership is only used for consumers from the same pod restarting?
    //
    // If set, allows static membership which would allow restarts within SESSION_TIMEOUT_MS_CONFIG
    // to happen with no rebalance

    return properties
}

fun KafkaProperties.withStreamThreads(n: Int): KafkaProperties {
    val properties = KafkaProperties()

    properties.putAll(this.toMap())
    properties[StreamsConfig.NUM_STREAM_THREADS_CONFIG] = n

    return properties
}

fun KafkaProperties.withErrorHandlers(deserializationExceptionHdl: DeserializationExceptionHandler?,
                                      streamExceptionHdl: StreamsUncaughtExceptionHandler?,
                                      productionExceptionHdl: ProductionExceptionHandler?): KafkaProperties {
    val properties = KafkaProperties()
    properties.putAll(this.toMap())

    deserializationExceptionHdl?.let  { properties[StreamsConfig.DEFAULT_DESERIALIZATION_EXCEPTION_HANDLER_CLASS_CONFIG] = it::class.java }
    streamExceptionHdl?.let           { properties[KAFKA_UNCAUGHT_EXCEPTION_HANDLER_CLASS_CONFIG] = it::class.java                        }
    productionExceptionHdl?.let       { properties[StreamsConfig.DEFAULT_PRODUCTION_EXCEPTION_HANDLER_CLASS_CONFIG] = it::class.java      }

    return properties
}
