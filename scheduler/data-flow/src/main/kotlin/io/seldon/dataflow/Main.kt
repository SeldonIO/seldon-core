/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow

import io.klogging.noCoLogger
import io.seldon.dataflow.kafka.*
import io.seldon.dataflow.kafka.security.KafkaSaslMechanisms
import io.seldon.dataflow.mtls.CertificateConfig
import io.seldon.dataflow.kafka.security.SaslConfig
import kotlinx.coroutines.runBlocking

object Main {
    private val logger = noCoLogger(Main::class)

    @JvmStatic
    fun main(args: Array<String>) {
        Logging.configure()

        val config = Cli.configWith(args)
        Logging.configure(
            appLevel = config[Cli.logLevelApplication],
            kafkaLevel = config[Cli.logLevelKafka],
        )

        val effectiveArgs = Cli.args().map { arg -> arg.name to config[arg] }
        logger.info { "initialised with config $effectiveArgs" }

        val tlsCertConfig = CertificateConfig(
            caCertPath = config[Cli.tlsCACertPath],
            keyPath = config[Cli.tlsKeyPath],
            certPath = config[Cli.tlsCertPath],
            brokerCaCertPath = config[Cli.brokerCACertPath],
            clientSecret = config[Cli.clientSecret],
            brokerSecret = config[Cli.brokerSecret],
            endpointIdentificationAlgorithm = config[Cli.endpointIdentificationAlgorithm],
        )

        val saslConfig = when (config[Cli.saslMechanism]) {
            KafkaSaslMechanisms.PLAIN -> SaslConfig.Password.Plain(
                secretName = config[Cli.saslSecret],
                username = config[Cli.saslUsername],
                passwordField = config[Cli.saslPasswordPath],
            )
            KafkaSaslMechanisms.SCRAM_SHA_256 -> SaslConfig.Password.Scram256(
                secretName = config[Cli.saslSecret],
                username = config[Cli.saslUsername],
                passwordField = config[Cli.saslPasswordPath],
            )
            KafkaSaslMechanisms.SCRAM_SHA_512 -> SaslConfig.Password.Scram512(
                secretName = config[Cli.saslSecret],
                username = config[Cli.saslUsername],
                passwordField = config[Cli.saslPasswordPath],
            )
            KafkaSaslMechanisms.OAUTH_BEARER -> SaslConfig.Oauth(
                secretName = config[Cli.saslSecret],
            )
        }

        val kafkaSecurityParams = KafkaSecurityParams(
            securityProtocol = config[Cli.kafkaSecurityProtocol],
            certConfig = tlsCertConfig,
            saslConfig = saslConfig,
        )
        val kafkaStreamsParams = KafkaStreamsParams(
            bootstrapServers = config[Cli.kafkaBootstrapServers],
            numPartitions = config[Cli.kafkaPartitions],
            replicationFactor = config[Cli.kafkaReplicationFactor],
            maxMessageSizeBytes = config[Cli.kafkaMaxMessageSizeBytes],
            security = kafkaSecurityParams,
        )
        val kafkaProperties = getKafkaProperties(kafkaStreamsParams)
        val kafkaAdminProperties = getKafkaAdminProperties(kafkaStreamsParams)
        val kafkaDomainParams = KafkaDomainParams(
            useCleanState = config[Cli.kafkaUseCleanState],
            joinWindowMillis = config[Cli.kafkaJoinWindowMillis],
        )
        val topicWaitRetryParams = TopicWaitRetryParams(
            createTimeoutMillis = config[Cli.topicCreateTimeoutMillis],
            describeTimeoutMillis = config[Cli.topicDescribeTimeoutMillis],
            describeRetries = config[Cli.topicDescribeRetries],
            describeRetryDelayMillis = config[Cli.topicDescribeRetryDelayMillis]
        )
        val subscriber = PipelineSubscriber(
            "seldon-dataflow-engine",
            kafkaProperties,
            kafkaAdminProperties,
            kafkaStreamsParams,
            kafkaDomainParams,
            topicWaitRetryParams,
            config[Cli.upstreamHost],
            config[Cli.upstreamPort],
            GrpcServiceConfigProvider.config,
            config[Cli.kafkaConsumerGroupIdPrefix],
            config[Cli.namespace],
        )

        addShutdownHandler(subscriber)

        runBlocking {
            subscriber.subscribe()
        }
    }

    private fun addShutdownHandler(subscriber: PipelineSubscriber) {
        Runtime.getRuntime().addShutdownHook(
            object : Thread() {
                override fun run() {
                    logger.info("received shutdown signal")
                    subscriber.cancelPipelines("shutting down")
                }
            }
        )
    }
}

// TODO - explore converting (sync?) KStreams into async Kotlin coroutines
