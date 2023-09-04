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

package io.seldon.dataflow.kafka

import io.seldon.dataflow.kafka.security.KafkaSaslMechanisms
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
import java.util.*

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

val kafkaTopicConfig = { maxMessageSizeBytes: Int ->
    mapOf(
        TopicConfig.MAX_MESSAGE_BYTES_CONFIG to maxMessageSizeBytes.toString(),
    )
}

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

        // Testing
        this[StreamsConfig.REPLICATION_FACTOR_CONFIG] = params.replicationFactor
        this[StreamsConfig.CACHE_MAX_BYTES_BUFFERING_CONFIG] = 0
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
    this[StreamsConfig.NUM_STREAM_THREADS_CONFIG] = n

    return properties
}
