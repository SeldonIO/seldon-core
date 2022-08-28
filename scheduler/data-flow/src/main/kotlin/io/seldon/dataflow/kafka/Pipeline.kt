package io.seldon.dataflow.kafka

import io.klogging.noCoLogger
import io.seldon.dataflow.hashutils.HashUtils
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineStepUpdate
import org.apache.kafka.streams.KafkaStreams
import org.apache.kafka.streams.KafkaStreams.StateListener
import org.apache.kafka.streams.StreamsBuilder
import org.apache.kafka.streams.Topology
import java.util.concurrent.CountDownLatch
import kotlin.math.floor
import kotlin.math.log2
import kotlin.math.max

typealias PipelineId = String

data class PipelineMetadata(
    val id: PipelineId,
    val name: String,
    val version: Int,
)

class Pipeline(
    private val metadata: PipelineMetadata,
    private val topology: Topology,
    private val streams: KafkaStreams,
    private val kafkaDomainParams: KafkaDomainParams,
    val size: Int,
) : StateListener {
    private val latch = CountDownLatch(1)

    fun start() {
        if (kafkaDomainParams.useCleanState) {
            streams.cleanUp()
        }
        logger.info {
            "starting pipeline ${metadata.name} (${metadata.id})" +
                    "\n" +
                    topology.describe()
        }
        streams.setStateListener(this)
        streams.start()

        // Do not allow pipeline to be marked as ready until it has successfully rebalanced.
        latch.await()
    }

    fun stop() {
        streams.close()
        // Does not clean up everything see https://issues.apache.org/jira/browse/KAFKA-13787
        streams.cleanUp()
    }

    override fun onChange(newState: KafkaStreams.State?, oldState: KafkaStreams.State?) {
        logger.info { "pipeline ${metadata.name} (v${metadata.version}) changing to state $newState" }
        if (newState == KafkaStreams.State.RUNNING) {
            latch.countDown()
        }
    }

    companion object {
        private val logger = noCoLogger(Pipeline::class)

        fun forSteps(
            metadata: PipelineMetadata,
            steps: List<PipelineStepUpdate>,
            kafkaProperties: KafkaProperties,
            kafkaDomainParams: KafkaDomainParams,
        ): Pipeline {
            val (topology, numSteps) = buildTopology(metadata, steps, kafkaDomainParams)
            val pipelineProperties = localiseKafkaProperties(kafkaProperties, metadata, numSteps)
            val streamsApp = KafkaStreams(topology, pipelineProperties)
            return Pipeline(metadata, topology, streamsApp, kafkaDomainParams, numSteps)
        }

        private fun buildTopology(
            metadata: PipelineMetadata,
            steps: List<PipelineStepUpdate>,
            kafkaDomainParams: KafkaDomainParams,
        ): Pair<Topology, Int> {
            val builder = StreamsBuilder()
            val topologySteps = steps
                .mapNotNull {
                    stepFor(
                        builder,
                        metadata.name,
                        it.sourcesList,
                        it.triggersList,
                        it.tensorMapMap,
                        it.sink,
                        it.inputJoinTy,
                        it.triggersJoinTy,
                        it.batch,
                        kafkaDomainParams,
                    )
                }
            val topology = builder.build()
            return topology to topologySteps.size
        }

        private fun localiseKafkaProperties(
            kafkaProperties: KafkaProperties,
            metadata: PipelineMetadata,
            numSteps: Int,
        ): KafkaProperties {
            return kafkaProperties
                .withAppId(
                    HashUtils.hashIfLong(metadata.id),
                )
                .withStreamThreads(
                    getNumThreadsFor(numSteps),
                )
        }

        private fun getNumThreadsFor(numSteps: Int): Int {
            val scale = floor(log2(numSteps.toFloat()))
            return max(1, scale.toInt())
        }
    }
}