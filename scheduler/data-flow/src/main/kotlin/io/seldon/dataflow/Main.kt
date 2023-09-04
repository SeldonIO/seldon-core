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

package io.seldon.dataflow

import io.klogging.noCoLogger
import io.seldon.dataflow.kafka.*
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

        val saslConfig = SaslConfig(
            mechanism = config[Cli.saslMechanism],
            username = config[Cli.saslUsername],
            credentialsSecret = config[Cli.saslSecret],
            passwordField = config[Cli.saslPasswordPath],
        )

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
        val subscriber = PipelineSubscriber(
            "seldon-dataflow-engine",
            kafkaProperties,
            kafkaAdminProperties,
            kafkaStreamsParams,
            kafkaDomainParams,
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
