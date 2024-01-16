/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow.kafka

import io.klogging.noCoLogger
import io.seldon.dataflow.hashutils.HashUtils
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineStepUpdate
import org.apache.kafka.streams.KafkaStreams
import org.apache.kafka.streams.KafkaStreams.StateListener
import org.apache.kafka.streams.StreamsBuilder
import org.apache.kafka.streams.StreamsConfig
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
    private var errorOrShutdown = false

    fun start() : Boolean {
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

        return !errorOrShutdown
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
            return
        }
        // CREATED, REBALANCING and RUNNING (with the latter one already handled above)
        // are the only non-error states. Everything else indicates an error or shutdown
        // and we should release the lock on which start() awaits and return an error.
        // see: https://kafka.apache.org/28/javadoc/org/apache/kafka/streams/KafkaStreams.State.html
        if (newState != KafkaStreams.State.CREATED && newState != KafkaStreams.State.REBALANCING) {
            errorOrShutdown = true
            latch.countDown()
        }

        // TODO: propagate pipeline state after initial startup (i.e. if the pipeline moves
        //       from RUNNING to REBALANCING or from RUNNING to an Error state)
    }

    companion object {
        private val logger = noCoLogger(Pipeline::class)

        fun forSteps(
            metadata: PipelineMetadata,
            steps: List<PipelineStepUpdate>,
            kafkaProperties: KafkaProperties,
            kafkaDomainParams: KafkaDomainParams,
            kafkaConsumerGroupIdPrefix: String,
            namespace: String,
        ): Pipeline {
            val (topology, numSteps) = buildTopology(metadata, steps, kafkaDomainParams)
            val pipelineProperties = localiseKafkaProperties(kafkaProperties, metadata, numSteps, kafkaConsumerGroupIdPrefix, namespace)
            val streamsApp = KafkaStreams(topology, pipelineProperties)
            logger.info("Create pipeline stream for name:${metadata.name} id:${metadata.id} version:${metadata.version} stream with kstream app id:${pipelineProperties[StreamsConfig.APPLICATION_ID_CONFIG]}")
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
                        it.tensorMapList,
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
            kafkaConsumerGroupIdPrefix: String,
            namespace: String
        ): KafkaProperties {
            return kafkaProperties
                .withAppId(
                    namespace,
                    kafkaConsumerGroupIdPrefix,
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