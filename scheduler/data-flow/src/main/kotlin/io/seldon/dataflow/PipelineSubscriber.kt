/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow

import com.github.michaelbull.retry.policy.binaryExponentialBackoff
import com.github.michaelbull.retry.retry
import io.grpc.ManagedChannelBuilder
import io.seldon.dataflow.kafka.CreationTask
import io.seldon.dataflow.kafka.DeletionTask
import io.seldon.dataflow.kafka.KafkaAdmin
import io.seldon.dataflow.kafka.KafkaAdminProperties
import io.seldon.dataflow.kafka.KafkaDomainParams
import io.seldon.dataflow.kafka.KafkaProperties
import io.seldon.dataflow.kafka.KafkaStreamsParams
import io.seldon.dataflow.kafka.Pipeline
import io.seldon.dataflow.kafka.PipelineId
import io.seldon.dataflow.kafka.PipelineMetadata
import io.seldon.dataflow.kafka.TopicWaitRetryParams
import io.seldon.dataflow.kafka.UpdateTask
import io.seldon.mlops.chainer.ChainerGrpcKt
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineStepUpdate
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineSubscriptionRequest
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateMessage
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateMessage.PipelineOperation
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateStatusMessage
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.FlowPreview
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.asCoroutineDispatcher
import kotlinx.coroutines.async
import kotlinx.coroutines.awaitAll
import kotlinx.coroutines.flow.collect
import kotlinx.coroutines.flow.onCompletion
import kotlinx.coroutines.flow.onEach
import kotlinx.coroutines.runBlocking
import java.util.concurrent.ConcurrentHashMap
import java.util.concurrent.Executors
import java.util.concurrent.TimeUnit
import io.klogging.logger as coLogger

@OptIn(FlowPreview::class)
class PipelineSubscriber(
    private val name: String,
    private val kafkaProperties: KafkaProperties,
    kafkaAdminProperties: KafkaAdminProperties,
    kafkaStreamsParams: KafkaStreamsParams,
    private val kafkaDomainParams: KafkaDomainParams,
    private val topicWaitRetryParams: TopicWaitRetryParams,
    private val upstreamHost: String,
    private val upstreamPort: Int,
    grpcServiceConfig: Map<String, Any>,
    private val kafkaConsumerGroupIdPrefix: String,
    private val namespace: String,
) {
    private val kafkaAdmin = KafkaAdmin(kafkaAdminProperties, kafkaStreamsParams, topicWaitRetryParams)
    private val channel =
        ManagedChannelBuilder
            .forAddress(upstreamHost, upstreamPort)
            .defaultServiceConfig(grpcServiceConfig)
            .usePlaintext() // Use TLS
            .enableRetry()
            // these keep alive settings need to match the go counterpart in scheduler/pkg/util/constants.go
            .keepAliveTime(60L, TimeUnit.SECONDS)
            .keepAliveTimeout(2L, TimeUnit.SECONDS)
            .build()

    val client = ChainerGrpcKt.ChainerCoroutineStub(channel)
    val pipelines = ConcurrentHashMap<PipelineId, Pipeline>()
    val dispatcher = Executors.newFixedThreadPool(20).asCoroutineDispatcher()
    val scope = CoroutineScope(SupervisorJob() + Dispatchers.Default)

    suspend fun subscribe() {
        while (true) {
            logger.info("will connect to $upstreamHost:$upstreamPort")
            retry(binaryExponentialBackoff(50..5_000L)) {
                logger.debug("retrying to connect to $upstreamHost:$upstreamPort")
                subscribePipelines(kafkaConsumerGroupIdPrefix, namespace)
            }
        }
    }

    // TODO
    //  - If a topology encounters an error, we should signal back to the scheduler about this.
    //  - If the scheduler updates/removes a topology, we need to cancel the corresponding coroutine.
    //  ...
    //  Pipeline UID should be enough to uniquely key it, even across versions?
    //  ...
    //  - Add map of model name -> (weak) referrents/reference count to avoid recreation of streams
    private suspend fun subscribePipelines(
        kafkaConsumerGroupIdPrefix: String,
        namespace: String,
    ) {
        logger.info("Subscribing to pipeline updates")
        client
            .subscribePipelineUpdates(request = makeSubscriptionRequest())
            .onEach { update ->
                logger.info("received request for ${update.pipeline}:${update.version} Id:${update.uid}")

                val metadata =
                    PipelineMetadata(
                        id = update.uid,
                        name = update.pipeline,
                        version = update.version,
                        pipelineOutputTopic = update.pipelineOutputTopic,
                        pipelineErrorTopic = update.pipelineErrorTopic,
                        allowCycles = update.allowCycles,
                        maxStepRevisits = update.maxStepRevisits,
                    )

                when (update.op) {
                    PipelineOperation.Create ->
                        handleCreate(
                            metadata,
                            update.updatesList,
                            update.timestamp,
                            kafkaConsumerGroupIdPrefix,
                            namespace,
                        )
                    PipelineOperation.Delete -> handleDelete(metadata, update.updatesList, update.timestamp)
                    else -> logger.warn("unrecognised pipeline operation (${update.op})")
                }
            }
            .onCompletion { cause ->
                if (cause == null) {
                    logger.info("pipeline subscription completed successfully")
                } else {
                    pipelines
                        .onEach {
                            // Defend against any existing pipelines that have failed but are not yet stopped, so that
                            // kafka streams may clean up resources (including temporary files). This is a catch-all
                            // and indicates we've missed calling stop in a failure case.
                            if (it.value.status.isError()) {
                                logger.debug(
                                    "(bug) pipeline in error state when subscription terminates with error. pipeline id: {pipelineId}",
                                    it.key,
                                )
                                it.value.stop()
                            }
                        }
                    logger.error("pipeline subscription terminated with error $cause")
                }
            }
            .collect()
        // TODO - use supervisor job(s) for spawning coroutines?
    }

    private fun makeSubscriptionRequest() =
        PipelineSubscriptionRequest
            .newBuilder()
            .setName(name)
            .build()

    private suspend fun handleCreate(
        metadata: PipelineMetadata,
        steps: List<PipelineStepUpdate>,
        timestamp: Long,
        kafkaConsumerGroupIdPrefix: String,
        namespace: String,
    ) {
        // If a pipeline with the same id exists, we assume it has the same name & version
        // If it's in an error state, try re-creating.
        //
        // WARNING: at the moment handleCreate is called sequentially on each update in
        // Flow<PipelineUpdateMessage> from subscribePipelines(). This allows us to sidestep issues
        // related to race conditions on `pipelines.containsKey(...)` below. If we ever move to
        // concurrent creation of pipelines, this needs to be revisited.
        if (pipelines.containsKey(metadata.id)) {
            val previous = pipelines[metadata.id]!!
            previous.queue.send(
                UpdateTask(
                    this,
                    metadata,
                    timestamp,
                    name,
                    logger,
                ),
            )
            return
        }

        // Create the new pipeline and start processing messages from
        // the task queue
        pipelines[metadata.id] =
            Pipeline(
                metadata,
                kafkaDomainParams,
                dispatcher,
                steps.size,
            ).also {
                it.startProcessing(scope)
            }

        // Send creation task to the task queue
        val pipeline = pipelines[metadata.id]!!
        pipeline.queue.send(
            CreationTask(
                this,
                metadata,
                steps,
                kafkaAdmin,
                kafkaProperties,
                kafkaDomainParams,
                kafkaConsumerGroupIdPrefix,
                namespace,
                timestamp,
                name,
                logger,
            ),
        )
    }

    private suspend fun handleDelete(
        metadata: PipelineMetadata,
        steps: List<PipelineStepUpdate>,
        timestamp: Long,
    ) {
        val pipeline = pipelines[metadata.id]
        if (pipeline != null) {
            // Remove pipeline from the subscriber so that future
            // creations of the pipeline will create a new entry in
            // the pipelines hashmap. This allows us to cleanly close
            // the task queue after deletion.
            pipelines.remove(metadata.id)

            // Send the deletion task. Note that we send a reference
            // to the pipeline to cleanly close the task queue
            pipeline.queue.send(
                DeletionTask(
                    pipeline,
                    this,
                    metadata,
                    steps,
                    kafkaAdmin,
                    timestamp,
                    name,
                    logger,
                ),
            )
        }
    }

    fun cancelPipelines(reason: String) {
        runBlocking {
            logger.info("cancelling pipelines due to: $reason")
            pipelines.values
                .map { pipeline ->
                    async { pipeline.stop() }
                }
                .awaitAll()
        }
    }

    fun makePipelineUpdateEvent(
        metadata: PipelineMetadata,
        operation: PipelineOperation,
        success: Boolean,
        reason: String = "",
        timestamp: Long,
        stream: String,
    ): PipelineUpdateStatusMessage {
        return PipelineUpdateStatusMessage
            .newBuilder()
            .setSuccess(success)
            .setReason(reason)
            .setUpdate(
                PipelineUpdateMessage
                    .newBuilder()
                    .setOp(operation)
                    .setPipeline(metadata.name)
                    .setVersion(metadata.version)
                    .setUid(metadata.id)
                    .setTimestamp(timestamp)
                    .setStream(stream)
                    .build(),
            )
            .build()
    }

    companion object {
        private val logger = coLogger(PipelineSubscriber::class)
    }
}
