package io.seldon.dataflow

import com.natpryce.konfig.*
import io.klogging.noCoLogger
import org.apache.kafka.common.security.auth.SecurityProtocol

object Cli {
    private const val envVarPrefix = "SELDON_"
    private val logger = noCoLogger(Cli::class)

    // Seldon components
    val upstreamHost = Key("upstream.host", stringType)
    val upstreamPort = Key("upstream.port", intType)

    // Kafka
    private val supportedKafkaProtocols = arrayOf(
        SecurityProtocol.PLAINTEXT,
        SecurityProtocol.SSL,
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

    fun configWith(rawArgs: Array<String>): Configuration {
        val fromProperties = ConfigurationProperties.fromResource("local.properties")
        val fromEnv = EnvironmentVariables(prefix = envVarPrefix)
        val fromArgs = parseArguments(rawArgs)

        return fromArgs overriding fromEnv overriding fromProperties
    }

    private fun parseArguments(rawArgs: Array<String>): Configuration {
        val (config, unparsedArgs) = parseArgs(
            rawArgs,
            CommandLineOption(kafkaBootstrapServers),
            CommandLineOption(kafkaSecurityProtocol),
            CommandLineOption(kafkaPartitions),
            CommandLineOption(kafkaReplicationFactor),
            CommandLineOption(kafkaUseCleanState),
            CommandLineOption(kafkaJoinWindowMillis),
            CommandLineOption(upstreamHost),
            CommandLineOption(upstreamPort),
            CommandLineOption(tlsCACertPath),
            CommandLineOption(tlsKeyPath),
            CommandLineOption(tlsCertPath),
            CommandLineOption(brokerCACertPath),
            CommandLineOption(clientSecret),
            CommandLineOption(brokerSecret),
            CommandLineOption(endpointIdentificationAlgorithm),
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