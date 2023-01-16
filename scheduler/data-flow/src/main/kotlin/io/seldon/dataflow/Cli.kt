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

import com.natpryce.konfig.*
import io.klogging.noCoLogger
import io.klogging.Level
import org.apache.kafka.common.security.auth.SecurityProtocol

object Cli {
    private const val envVarPrefix = "SELDON_"
    private val logger = noCoLogger(Cli::class)

    // General setup
    val logLevelApplication = Key("log.level.app", enumType(*Level.values()))
    val logLevelKafka = Key("log.level.kafka", enumType(*Level.values()))

    // Seldon components
    val upstreamHost = Key("upstream.host", stringType)
    val upstreamPort = Key("upstream.port", intType)

    // Kafka
    private val supportedKafkaProtocols = arrayOf(
        SecurityProtocol.PLAINTEXT,
        SecurityProtocol.SSL,
        SecurityProtocol.SASL_SSL,
    ) // TODO - move to Kafka package
    val kafkaBootstrapServers = Key("kafka.bootstrap.servers", stringType)
    val kafkaSecurityProtocol = Key("kafka.security.protocol", enumType(*supportedKafkaProtocols))
    val kafkaPartitions = Key("kafka.partitions.default", intType)
    val kafkaReplicationFactor = Key("kafka.replication.factor", intType)
    val kafkaUseCleanState = Key("kafka.state.clean", booleanType)
    val kafkaJoinWindowMillis = Key("kafka.join.window.millis", longType)

    // Mutual TLS
    val tlsCACertPath = Key("tls.client.ca.path", stringType)
    val tlsKeyPath = Key("tls.client.key.path", stringType)
    val tlsCertPath = Key("tls.client.cert.path", stringType)
    val brokerCACertPath = Key("tls.broker.ca.path", stringType)
    val clientSecret = Key("tls.client.secret", stringType)
    val brokerSecret = Key("tls.broker.secret", stringType)
    val endpointIdentificationAlgorithm = Key("tls.endpoint.identification.algorithm", stringType)

    // SASL
    val saslUsername = Key("sasl.username", stringType)
    val saslSecret = Key("sasl.secret", stringType)
    val saslPasswordPath = Key("sasl.password.path", stringType)

    fun args(): List<Key<Any>> {
        return listOf(
            logLevelApplication,
            logLevelKafka,
            upstreamHost,
            upstreamPort,
            kafkaBootstrapServers,
            kafkaSecurityProtocol,
            kafkaPartitions,
            kafkaReplicationFactor,
            kafkaUseCleanState,
            kafkaJoinWindowMillis,
            tlsCACertPath,
            tlsKeyPath,
            tlsCertPath,
            brokerCACertPath,
            clientSecret,
            brokerSecret,
            endpointIdentificationAlgorithm,
            saslUsername,
            saslSecret,
            saslPasswordPath
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