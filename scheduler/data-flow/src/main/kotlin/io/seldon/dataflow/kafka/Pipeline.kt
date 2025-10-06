/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow.kafka

import io.klogging.noCoLogger
import io.seldon.dataflow.PipelineSubscriber
import io.seldon.dataflow.hashutils.HashUtils
import io.seldon.dataflow.withException
import io.seldon.dataflow.withMessage
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineStepUpdate
import kotlinx.coroutines.runBlocking
import org.apache.kafka.common.serialization.Serdes
import org.apache.kafka.streams.KafkaStreams
import org.apache.kafka.streams.KafkaStreams.State
import org.apache.kafka.streams.KafkaStreams.StateListener
import org.apache.kafka.streams.StreamsBuilder
import org.apache.kafka.streams.StreamsConfig
import org.apache.kafka.streams.Topology
import org.apache.kafka.streams.errors.StreamsUncaughtExceptionHandler
import org.apache.kafka.streams.state.Stores
import java.util.concurrent.CountDownLatch
import java.util.concurrent.locks.Lock
import java.util.concurrent.locks.ReentrantLock
import kotlin.math.floor
import kotlin.math.log2
import kotlin.math.max

inline fun <T> Lock.withLock(action: () -> T): T {
    lock()
    try {
        return action()
    } finally {
        unlock()
    }
}

typealias PipelineId = String

data class PipelineMetadata(
    val id: PipelineId,
    val name: String,
    val version: Int,
    val pipelineOutputTopic: String,
    val pipelineErrorTopic: String,
    val allowCycles: Boolean,
    val maxStepRevisits: Int,
)

class Pipeline(
    private val metadata: PipelineMetadata,
    private val topology: Topology,
    private val streams: KafkaStreams,
    private val kafkaDomainParams: KafkaDomainParams,
    val size: Int,
    val pipelineSubscriber: PipelineSubscriber,
    var timestamp: Long,
) : StateListener {
    private val latch = CountDownLatch(1)
    private val statusLock = ReentrantLock()

    // Never update status properties in-place, because we need it to have atomic
    // properties. Instead, just assign new values to it.
    @Volatile
    var status: PipelineStatus = PipelineStatus.StreamStopped(null)

    fun start(): PipelineStatus {
        if (kafkaDomainParams.useCleanState) {
            streams.cleanUp()
        }
        logger.info("starting pipeline {pipelineName}, ({pipelineId}, {pipelineVersion})", metadata.name, metadata.id, metadata.version)
        logger.debug { topology.describe() }
        streams.setStateListener(this)

        try {
            statusLock.withLock {
                streams.start()
                status = PipelineStatus.StreamStarting()
            }
        } catch (e: Exception) {
            statusLock.withLock {
                streams.close()
                streams.cleanUp()
                status =
                    PipelineStatus.Error(State.NOT_RUNNING)
                        .withException(e)
                        .withMessage("kafka streams: failed to start")
            }
            return status
        }

        // Wait until the pipeline is successfully rebalanced or reaches an error state.
        // The pipeline status will be updated by the onChange callback, or by the stop() function
        //
        // Note: stop() may be called asynchronously at any point, if the process is signaled
        // to stop. Therefore, it possible for the start function to return with status
        // PipelineStatus.StreamStopped.
        latch.await()
        return status
    }

    fun stop() {
        var prevStatus: PipelineStatus? = null
        statusLock.withLock {
            prevStatus = status
            status = PipelineStatus.StreamStopping()
        }

        // Close needs to be called even if streams.start() never gets called
        streams.close()
        // Does not clean up everything see https://issues.apache.org/jira/browse/KAFKA-13787
        streams.cleanUp()
        statusLock.withLock {
            status = PipelineStatus.StreamStopped(prevStatus)
        }

        // if stop() is called while start() is still waiting on the latch, release it so that
        // it may return the stopped status
        latch.countDown()
    }

    override fun onChange(
        newState: State?,
        oldState: State?,
    ) {
        logger.info {
            e(
                "pipeline {pipelineName} (v{pipelineVersion}) changing to state $newState",
                metadata.name,
                metadata.version,
            )
        }

        statusLock.withLock {
            if (status is PipelineStatus.StreamStopping) {
                return
            }

            if (latch.count == 1L) {
                // count == 1 means that we are in the create phase of the pipeline
                //
                // Only update the status if the pipeline is not already being stopped
                // we wouldn't want the status to transition to Started after it was
                // marked as StreamStopping. If status is StreamStopping, we guarantee
                // that latch.countDown() will be called.
                status =
                    when (newState) {
                        State.RUNNING -> PipelineStatus.Started()
                        State.CREATED, State.REBALANCING -> status

                        // CREATED, REBALANCING and RUNNING are the only non-error states. Everything
                        // else indicates an error or shutdown, and we should release the lock on which
                        // start() awaits and return an error.
                        // see: https://kafka.apache.org/28/javadoc/org/apache/kafka/streams/KafkaStreams.State.html
                        else ->
                            PipelineStatus.Error(newState)
                                .withMessage("pipeline data streams error: kafka streams state: $newState")
                    }

                // Release the latch when the pipeline is running, or it encountered and error
                if (newState != State.CREATED && newState != State.REBALANCING) {
                    latch.countDown()
                }
            } else {
                // Pipeline is running, and it transitioned from RUNNING -> REBALANCING
                // or from REBALANCING -> RUNNING. Other events must be errors since we
                // handle pipeline shutdown in the stop function
                status =
                    when (newState) {
                        State.RUNNING -> PipelineStatus.Started()
                        State.REBALANCING -> PipelineStatus.StreamRebalancing()
                        else ->
                            PipelineStatus.Error(newState)
                                .withMessage("pipeline data streams error: kafka streams state: $newState")
                    }

                // Send the new status to the scheduler.
                runBlocking {
                    pipelineSubscriber.handleUpdate(
                        metadata = metadata,
                        timestamp = timestamp,
                        status = status,
                    )
                }
            }
        }
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
            kafkaStreamsSerdes: KafkaStreamsSerdes,
            pipelineSubscriber: PipelineSubscriber,
            timestamp: Long,
        ): Pair<Pipeline?, PipelineStatus.Error?> {
            val (topology, numSteps) = buildTopology(metadata, steps, kafkaDomainParams, kafkaStreamsSerdes)
            val pipelineProperties = localiseKafkaProperties(kafkaProperties, metadata, numSteps, kafkaConsumerGroupIdPrefix, namespace)
            var streamsApp: KafkaStreams?
            var pipelineError: PipelineStatus.Error?
            try {
                streamsApp = KafkaStreams(topology, pipelineProperties)
            } catch (e: Exception) {
                pipelineError =
                    PipelineStatus.Error(null)
                        .withException(e)
                        .withMessage("failed to initialize kafka streams for pipeline")
                return null to pipelineError
            }

            val uncaughtExceptionHandlerClass =
                pipelineProperties[KAFKA_UNCAUGHT_EXCEPTION_HANDLER_CLASS_CONFIG] as? Class<StreamsUncaughtExceptionHandler>?
            uncaughtExceptionHandlerClass?.let {
                logger.debug("Setting custom Kafka streams uncaught exception handler")
                streamsApp.setUncaughtExceptionHandler(it.getDeclaredConstructor().newInstance())
            }
            logger.info(
                "Create pipeline stream for name:{pipelineName} id:{pipelineId} " +
                    "version:{pipelineVersion} stream with kstream app id:{kstreamAppId}",
                metadata.name,
                metadata.id,
                metadata.version,
                pipelineProperties[StreamsConfig.APPLICATION_ID_CONFIG],
            )
            logger.info(
                "AllowCycles: ${metadata.allowCycles}; maxStepRevisits: ${metadata.maxStepRevisits}",
            )
            return Pipeline(metadata, topology, streamsApp, kafkaDomainParams, numSteps, pipelineSubscriber, timestamp) to null
        }

        private fun buildTopology(
            metadata: PipelineMetadata,
            steps: List<PipelineStepUpdate>,
            kafkaDomainParams: KafkaDomainParams,
            kafkaStreamsSerdes: KafkaStreamsSerdes,
        ): Pair<Topology, Int> {
            // Sort ensure the same order when building the topology amongst multiple
            // replicas. The scheduler doesn't send the same message because the steps
            // are created from iterating over a map
            val sortedSteps =
                steps.sortedWith(
                    compareBy(
                        { step -> step.sink.topicName },
                        { step -> step.sourcesList.joinToString(",") { it.topicName } },
                        { step -> step.triggersList.joinToString(",") { it.topicName } },
                    ),
                )

            val builder = StreamsBuilder()

            if (metadata.allowCycles) {
                builder.addStateStore(
                    Stores.keyValueStoreBuilder(
                        Stores.inMemoryKeyValueStore(VISITING_COUNTER_STORE),
                        Serdes.String(),
                        Serdes.Integer(),
                    ),
                )
            }

            val topologySteps =
                sortedSteps
                    .mapNotNull {
                        stepFor(
                            builder,
                            metadata.name,
                            metadata.version.toString(),
                            metadata.pipelineOutputTopic,
                            metadata.pipelineErrorTopic,
                            metadata.allowCycles,
                            metadata.maxStepRevisits,
                            it.sourcesList,
                            it.triggersList,
                            it.tensorMapList,
                            it.sink,
                            it.inputJoinTy,
                            it.triggersJoinTy,
                            it.joinWindowMs.toLong(),
                            it.batch,
                            kafkaDomainParams,
                            kafkaStreamsSerdes,
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
            namespace: String,
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
                .withErrorHandlers(
                    StreamErrorHandling.StreamsDeserializationErrorHandler(),
                    StreamErrorHandling.StreamsCustomUncaughtExceptionHandler(),
                    StreamErrorHandling.StreamsRecordProducerErrorHandler(),
                )
        }

        private fun getNumThreadsFor(numSteps: Int): Int {
            val scale = floor(log2(numSteps.toFloat()))
            return max(1, scale.toInt())
        }
    }
}
