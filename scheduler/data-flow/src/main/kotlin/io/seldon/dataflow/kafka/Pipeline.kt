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
import io.seldon.dataflow.withException
import io.seldon.dataflow.withMessage
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineStepUpdate
import org.apache.kafka.streams.KafkaStreams
import org.apache.kafka.streams.KafkaStreams.State
import org.apache.kafka.streams.KafkaStreams.StateListener
import org.apache.kafka.streams.StreamsBuilder
import org.apache.kafka.streams.StreamsConfig
import org.apache.kafka.streams.Topology
import org.apache.kafka.streams.errors.StreamsException
import org.apache.kafka.streams.errors.StreamsUncaughtExceptionHandler
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
    // Never update status properties in-place, because we need it to have atomic
    // properties. Instead, just assign new values to it.
    @Volatile
    var status : PipelineStatus = PipelineStatus.StreamStopped(null)

    fun start() : PipelineStatus {
        if (kafkaDomainParams.useCleanState) {
            streams.cleanUp()
        }
        logger.info("starting pipeline {pipelineName}, ({pipelineId})", metadata.name, metadata.id)
        logger.debug { topology.describe() }
        streams.setStateListener(this)
        try {
            streams.start()
        } catch (e: Exception) {
            streams.close()
            streams.cleanUp()
            status = PipelineStatus.Error(State.NOT_RUNNING)
                .withException(e)
                .withMessage("kafka streams: failed to start")
            return status
        }
        status = PipelineStatus.StreamStarting()

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
        val prevStatus = status
        status = PipelineStatus.StreamStopping()
        // Close needs to be called even if streams.start() never gets called
        streams.close()
        // Does not clean up everything see https://issues.apache.org/jira/browse/KAFKA-13787
        streams.cleanUp()
        status = PipelineStatus.StreamStopped(prevStatus)

        // if stop() is called while start() is still waiting on the latch, release it so that
        // it may return the stopped status
        latch.countDown()
    }

    override fun onChange(newState: State?, oldState: State?) {
        logger.info {
            e("pipeline {pipelineName} (v{pipelineVersion}) changing to state $newState",
                metadata.name,
                metadata.version)
        }
        if (newState == State.RUNNING) {
            // Only update the status if the pipeline is not already being stopped
            // we wouldn't want the status to transition to Started after it was
            // marked as StreamStopping. If status is StreamStopping, we guarantee
            // that latch.countDown() will be called.
            if (status !is PipelineStatus.StreamStopping) {
                status = PipelineStatus.Started()
                latch.countDown()
            }

            return
        }
        // CREATED, REBALANCING and RUNNING (with the latter one already handled above)
        // are the only non-error states. Everything else indicates an error or shutdown
        // and we should release the lock on which start() awaits and return an error.
        // see: https://kafka.apache.org/28/javadoc/org/apache/kafka/streams/KafkaStreams.State.html
        if (newState != State.CREATED && newState != State.REBALANCING) {
            if (status !is PipelineStatus.StreamStopping) {
                status = PipelineStatus.Error(newState)
                    .withMessage("pipeline data streams error: kafka streams state: $newState")
                latch.countDown()
            }
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
        ): Pair<Pipeline?, PipelineStatus.Error?> {
            val (topology, numSteps) = buildTopology(metadata, steps, kafkaDomainParams)
            val pipelineProperties = localiseKafkaProperties(kafkaProperties, metadata, numSteps, kafkaConsumerGroupIdPrefix, namespace)
            var streamsApp : KafkaStreams? = null
            var pipelineError: PipelineStatus.Error? = null
            try {
                streamsApp = KafkaStreams(topology, pipelineProperties)
            } catch (e: StreamsException) {
                pipelineError = PipelineStatus.Error(null)
                    .withException(e)
                    .withMessage("failed to initialize kafka streams app")
                return null to pipelineError
            }

            val uncaughtExceptionHandlerClass = pipelineProperties[KAFKA_UNCAUGHT_EXCEPTION_HANDLER_CLASS_CONFIG] as? Class<StreamsUncaughtExceptionHandler>?
            uncaughtExceptionHandlerClass?.let{
                logger.info("Setting custom Kafka streams uncaught exception handler")
                streamsApp.setUncaughtExceptionHandler(it.getDeclaredConstructor().newInstance())
            }
            logger.info(
                "Create pipeline stream for name:{pipelineName} id:{pipelineId} version:{pipelineVersion} stream with kstream app id:{kstreamAppId}",
                metadata.name,
                metadata.id,
                metadata.version,
                pipelineProperties[StreamsConfig.APPLICATION_ID_CONFIG]
            )
            return Pipeline(metadata, topology, streamsApp, kafkaDomainParams, numSteps) to null
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
                .withErrorHandlers(
                    StreamErrorHandling.StreamsDeserializationErrorHandler(),
                    StreamErrorHandling.StreamsCustomUncaughtExceptionHandler(),
                    StreamErrorHandling.StreamsRecordProducerErrorHandler()
                )
        }

        private fun getNumThreadsFor(numSteps: Int): Int {
            val scale = floor(log2(numSteps.toFloat()))
            return max(1, scale.toInt())
        }
    }
}