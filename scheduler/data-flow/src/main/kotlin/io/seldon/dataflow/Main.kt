package io.seldon.dataflow

import io.klogging.noCoLogger
import io.seldon.dataflow.kafka.KafkaParams
import io.seldon.dataflow.kafka.getKafkaProperties
import kotlinx.coroutines.runBlocking

object Main {
    private val logger = noCoLogger(Main::class)

    @JvmStatic
    fun main(args: Array<String>) {
        Logging.configure()

        val config = Cli.configWith(args)
        logger.info("initialised")

        val kafkaProperties = getKafkaProperties(
            KafkaParams(
                bootstrapServers = config[Cli.bootstrapServers],
                numCores = config[Cli.numCores],
            ),
        )
        val subscriber = PipelineSubscriber(
            "seldon-dataflow-transformer",
            kafkaProperties,
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

// Steps:
// - Should not create any topics - just do joins (check Kafka settings to disable auto-creation)
//      - Can log warning and retry
// - Know how to parse source/sink names (seldon.<...>.<uid?>)
// - Read input topic(s), transform into single output message (maybe many tensors), write to output topic
// - Add gRPC client to connect to scheduler
// - Respond to management calls over protos - start/stop handling topics

// TODO - explore converting (sync?) KStreams into async Kotlin coroutines