/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow

import com.charleskorn.kaml.Yaml
import com.natpryce.konfig.CommandLineOption
import com.natpryce.konfig.Configuration
import com.natpryce.konfig.ConfigurationMap
import com.natpryce.konfig.ConfigurationProperties
import com.natpryce.konfig.EnvironmentVariables
import com.natpryce.konfig.Key
import com.natpryce.konfig.booleanType
import com.natpryce.konfig.enumType
import com.natpryce.konfig.intType
import com.natpryce.konfig.longType
import com.natpryce.konfig.overriding
import com.natpryce.konfig.parseArgs
import com.natpryce.konfig.stringType
import io.klogging.Level
import io.klogging.noCoLogger
import io.seldon.dataflow.kafka.SchemaRegistryConfig
import io.seldon.dataflow.kafka.security.KafkaSaslMechanisms
import io.seldon.dataflow.kafka.security.KafkaSecurityProtocols
import kotlinx.serialization.Serializable
import java.io.File
import java.net.InetAddress
import java.util.UUID

@Serializable
data class ConfluentSchemaConfig(
    val schemaRegistry: SchemaRegistrySection = SchemaRegistrySection(),
) {
    @Serializable
    data class SchemaRegistrySection(
        val client: ClientConfig = ClientConfig(),
    )

    @Serializable
    data class ClientConfig(
        val URL: String = "",
        val username: String = "",
        val password: String = "",
    )
}

object Cli {
    private const val ENV_VAR_PREFIX = "SELDON_"
    private const val CONFLUENTSCHEMAFILENAME = ".confluent-schema.yaml"
    private val logger = noCoLogger(Cli::class)

    // General setup
    val logLevelApplication = Key("log.level.app", enumType(*Level.values()))
    val logLevelKafka = Key("log.level.kafka", enumType(*Level.values()))
    val namespace = Key("pod.namespace", stringType)
    val dataflowReplicaId = Key("dataflow.replica.id", stringType)
    val pipelineCtlopsThreads = Key("pipeline.ctlops.threads", intType)

    // Health probe server
    val healthServerPort = Key("health.server.port", intType)

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

    // Schema Registry
    val schemaRegistryConfigPath = Key("kafka.schema.registry.config.path", stringType)

    fun args(): List<Key<Any>> {
        return listOf(
            logLevelApplication,
            logLevelKafka,
            namespace,
            dataflowReplicaId,
            pipelineCtlopsThreads,
            healthServerPort,
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
            schemaRegistryConfigPath,
        )
    }

    fun getSchemaConfig(config: Configuration): SchemaRegistryConfig {
        if (config[schemaRegistryConfigPath] == "") {
            logger.debug("not using schema registry")
            return SchemaRegistryConfig(_useSchemaRegistry = false)
        }

        val configPath = config[schemaRegistryConfigPath]
        val schemaConfigFile = File(configPath, CONFLUENTSCHEMAFILENAME)

        if (!schemaConfigFile.exists()) {
            val errorMsg = "Schema config file not found at: ${schemaConfigFile.absolutePath}"
            logger.error(errorMsg)
            throw IllegalStateException(errorMsg)
        }

        return try {
            val yamlContent = schemaConfigFile.readText()
            val confluentConfig = Yaml.default.decodeFromString(ConfluentSchemaConfig.serializer(), yamlContent)
            logger.info("read config file for schema registry")

            val clientConfig = confluentConfig.schemaRegistry.client
            logger.info("the config is URL=${clientConfig.URL}, username=${clientConfig.username}")
            val schemaConfig =
                SchemaRegistryConfig(
                    url = clientConfig.URL,
                    username = clientConfig.username,
                    password = clientConfig.password,
                )
            schemaConfig.validate()
            schemaConfig
        } catch (e: Exception) {
            val errorMsg = "Failed to load or validate schema config from: ${schemaConfigFile.absolutePath}"
            logger.error(errorMsg, e)
            throw IllegalStateException(errorMsg, e)
        }
    }

    fun configWith(rawArgs: Array<String>): Configuration {
        val fromProperties = ConfigurationProperties.fromResource("local.properties")
        val fromSystem = getSystemConfig()
        val fromEnv = EnvironmentVariables(prefix = ENV_VAR_PREFIX)
        val fromArgs = parseArguments(rawArgs)

        return fromArgs overriding fromEnv overriding fromSystem overriding fromProperties
    }

    private fun getSystemConfig(): Configuration {
        val dataflowIdPair = this.dataflowReplicaId to getNewDataflowId()
        return ConfigurationMap(dataflowIdPair)
    }

    fun getNewDataflowId(assignRandomUuid: Boolean = false): String {
        if (!assignRandomUuid) {
            try {
                return InetAddress.getLocalHost().hostName
            } catch (_: Exception) {
            }
        }
        return "seldon-dataflow-engine-" + UUID.randomUUID().toString()
    }

    private fun parseArguments(rawArgs: Array<String>): Configuration {
        val (config, unparsedArgs) =
            parseArgs(
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
