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
import javax.xml.stream.events.Namespace
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
        private val namespace = System.getenv("POD_NAMESPACE")

        fun forSteps(
            metadata: PipelineMetadata,
            steps: List<PipelineStepUpdate>,
            kafkaProperties: KafkaProperties,
            kafkaDomainParams: KafkaDomainParams,
            kafkaConsumerGroupIdPrefix: String,
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