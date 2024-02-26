/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow

import com.natpryce.konfig.*
import io.klogging.Level
import io.klogging.noCoLogger
import io.seldon.dataflow.kafka.security.KafkaSaslMechanisms
import io.seldon.dataflow.kafka.security.KafkaSecurityProtocols

object Cli {
    private const val envVarPrefix = "SELDON_"
    private val logger = noCoLogger(Cli::class)

    // General setup
    val logLevelApplication = Key("log.level.app", enumType(*Level.values()))
    val logLevelKafka = Key("log.level.kafka", enumType(*Level.values()))
    val namespace = Key("pod.namespace", stringType)

    // Seldon components
    val upstreamHost = Key("upstream.host", stringType)
    val upstreamPort = Key("upstream.port", intType)

    // Kafka
    val kafkaBootstrapServers = Key("kafka.bootstrap.servers", stringType)
    val kafkaConsumerGroupIdPrefix = Key("kafka.consumer.prefix", stringType)
    val kafkaSecurityProtocol = Key("kafka.security.protocol", enumType(*KafkaSecurityProtocols))
    val kafkaPartitions = Key("kafka.partitions.default", intType)
    val kafkaReplicationFactor = Key("kafka.replication.factor", intType)
    val kafkaUseCleanState = Key("kafka.state.clean", booleanType)
    val kafkaJoinWindowMillis = Key("kafka.join.window.millis", longType)
    val kafkaMaxMessageSizeBytes = Key("kafka.max.message.size.bytes", intType)

    // Kafka (m)TLS
    val tlsCACertPath = Key("kafka.tls.client.ca.path", stringType)
    val tlsKeyPath = Key("kafka.tls.client.key.path", stringType)
    val tlsCertPath = Key("kafka.tls.client.cert.path", stringType)
    val brokerCACertPath = Key("kafka.tls.broker.ca.path", stringType)
    val clientSecret = Key("kafka.tls.client.secret", stringType)
    val brokerSecret = Key("kafka.tls.broker.secret", stringType)
    val endpointIdentificationAlgorithm = Key("kafka.tls.endpoint.identification.algorithm", stringType)

    // Kafka waiting for topic creation
    val topicCreateTimeoutMillis = Key("topic.create.timeout.millis", intType)
    val topicDescribeTimeoutMillis = Key("topic.describe.timeout.millis", longType)
    val topicDescribeRetries = Key("topic.describe.retry.attempts", intType)
    val topicDescribeRetryDelayMillis = Key("topic.describe.retry.delay.millis", longType)

    // Kafka SASL
    val saslUsername = Key("kafka.sasl.username", stringType)
    val saslSecret = Key("kafka.sasl.secret", stringType)
    val saslPasswordPath = Key("kafka.sasl.password.path", stringType)
    val saslMechanism = Key("kafka.sasl.mechanism", enumType(KafkaSaslMechanisms.byName))

    fun args(): List<Key<Any>> {
        return listOf(
            logLevelApplication,
            logLevelKafka,
            namespace,
            upstreamHost,
            upstreamPort,
            kafkaBootstrapServers,
            kafkaConsumerGroupIdPrefix,
            kafkaSecurityProtocol,
            kafkaPartitions,
            kafkaReplicationFactor,
            kafkaUseCleanState,
            kafkaJoinWindowMillis,
            kafkaMaxMessageSizeBytes,
            tlsCACertPath,
            tlsKeyPath,
            tlsCertPath,
            brokerCACertPath,
            clientSecret,
            brokerSecret,
            endpointIdentificationAlgorithm,
            topicCreateTimeoutMillis,
            topicDescribeTimeoutMillis,
            topicDescribeRetries,
            topicDescribeRetryDelayMillis,
            saslUsername,
            saslSecret,
            saslPasswordPath,
            saslMechanism,
        )
    }

    fun configWith(rawArgs: Array<String>): Configuration {
        val fromProperties = ConfigurationProperties.fromResource("local.properties")
        val fromEnv = EnvironmentVariables(prefix = envVarPrefix)
        val fromArgs = parseArguments(rawArgs)

        return fromArgs overriding fromEnv overriding fromProperties
    }

    private fun parseArguments(rawArgs: Array<String>): Configuration {
        val (config, unparsedArgs) = parseArgs(
            rawArgs,
            *this.args().map { CommandLineOption(it) }.toTypedArray(),
            programName = "seldon-dataflow-engine",
        )
        if (unparsedArgs.isNotEmpty()) {
            logUnknownArguments(unparsedArgs)
        }
        return config
    }

    private fun logUnknownArguments(unknownArgs: List<String>) {
        logger.warn(
            "received unexpected arguments: {unknownArgs}",
            unknownArgs,
        )
    }
}
