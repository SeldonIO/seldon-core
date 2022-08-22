package io.seldon.dataflow

import io.klogging.noCoLogger
import io.seldon.dataflow.kafka.KafkaDomainParams
import io.seldon.dataflow.kafka.KafkaStreamsParams
import io.seldon.dataflow.kafka.getKafkaAdminProperties
import io.seldon.dataflow.kafka.getKafkaProperties
import kotlinx.coroutines.runBlocking

object Main {
    private val logger = noCoLogger(Main::class)

    @JvmStatic
    fun main(args: Array<String>) {
        Logging.configure()

        val config = Cli.configWith(args)
        logger.info("initialised")

        val kafkaStreamsParams = KafkaStreamsParams(
            bootstrapServers = config[Cli.kafkaBootstrapServers],
            securityProtocol = config[Cli.kafkaSecurityProtocol],
            numPartitions = config[Cli.kafkaPartitions],
            replicationFactor = config[Cli.kafkaReplicationFactor],
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