package io.seldon.dataflow

import com.natpryce.konfig.*
import io.klogging.noCoLogger

object Cli {
    private const val envVarPrefix = "SELDON"
    private val logger = noCoLogger(Cli::class)

    val bootstrapServers = Key("kafka.bootstrap.servers", stringType)
    val numCores = Key("cores.count", intType)
    val upstreamHost = Key("upstream.host", stringType)
    val upstreamPort = Key("upstream.port", intType)

    fun configWith(rawArgs: Array<String>): Configuration {
        val fromProperties = ConfigurationProperties.fromResource("local.properties")
        val fromEnv = EnvironmentVariables(prefix = envVarPrefix)
        val fromArgs = parseArguments(rawArgs)

        return fromArgs overriding fromEnv overriding fromProperties
    }

    private fun parseArguments(rawArgs: Array<String>): Configuration {
        val (config, unparsedArgs) = parseArgs(
            rawArgs,
            CommandLineOption(bootstrapServers),
            CommandLineOption(numCores),
            CommandLineOption(upstreamHost),
            CommandLineOption(upstreamPort),
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