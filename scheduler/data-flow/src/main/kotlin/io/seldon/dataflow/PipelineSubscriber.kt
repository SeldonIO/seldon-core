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
import io.klogging.Level
import io.seldon.dataflow.kafka.KafkaAdmin
import io.seldon.dataflow.kafka.KafkaAdminProperties
import io.seldon.dataflow.kafka.KafkaDomainParams
import io.seldon.dataflow.kafka.KafkaProperties
import io.seldon.dataflow.kafka.KafkaStreamsParams
import io.seldon.dataflow.kafka.Pipeline
import io.seldon.dataflow.kafka.PipelineId
import io.seldon.dataflow.kafka.PipelineMetadata
import io.seldon.dataflow.kafka.PipelineStatus
import io.seldon.dataflow.kafka.TopicWaitRetryParams
import io.seldon.mlops.chainer.ChainerGrpcKt
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineStepUpdate
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineSubscriptionRequest
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateMessage
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateMessage.PipelineOperation
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateStatusMessage
import kotlinx.coroutines.FlowPreview
import kotlinx.coroutines.async
import kotlinx.coroutines.awaitAll
import kotlinx.coroutines.flow.collect
import kotlinx.coroutines.flow.onCompletion
import kotlinx.coroutines.flow.onEach
import kotlinx.coroutines.runBlocking
import java.util.concurrent.ConcurrentHashMap
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
            .build()
    private val client = ChainerGrpcKt.ChainerCoroutineStub(channel)

    private val pipelines = ConcurrentHashMap<PipelineId, Pipeline>()

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
                    )

                when (update.op) {
                    PipelineOperation.Create -> handleCreate(metadata, update.updatesList, kafkaConsumerGroupIdPrefix, namespace)
                    PipelineOperation.Delete -> handleDelete(metadata)
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
        kafkaConsumerGroupIdPrefix: String,
        namespace: String,
    ) {
        val defaultReason = "pipeline created"
        // If a pipeline with the same id exists, we assume it has the same name & version
        // If it's in an error state, try re-creating.
        //
        // WARNING: at the moment handleCreate is called sequentially on each update in
        // Flow<PipelineUpdateMessage> from subscribePipelines(). This allows us to sidestep issues
        // related to race conditions on `pipelines.containsKey(...)` below. If we ever move to
        // concurrent creation of pipelines, this needs to be revisited.
        if (pipelines.containsKey(metadata.id)) {
            val previous = pipelines[metadata.id]!!
            if (previous.status.isActive()) {
                client.pipelineUpdateEvent(
                    makePipelineUpdateEvent(
                        metadata = metadata,
                        operation = PipelineOperation.Create,
                        success = true,
                        reason = previous.status.getDescription() ?: defaultReason,
                    ),
                )
                logger.debug(
                    "response to scheduler: pipeline {pipelineName} continues to run normally; " +
                        "pipeline version: {pipelineVersion}, id: {pipelineId}",
                    metadata.name,
                    metadata.version,
                    metadata.id,
                )
                return
            } else { // pipeline exists but in failed/stopped state; cleanup state and re-create
                logger.info(
                    "Recreating non-active pipeline {pipelineName} version: {pipelineVersion}, id: {pipelineId}",
                    metadata.name,
                    metadata.version,
                    metadata.id,
                )
                logger.debug(
                    "Previous state for non-active pipeline {pipelineName} version: {pipelineVersion}, id: {pipelineId}: {pipelineStatus}",
                    metadata.name,
                    metadata.version,
                    metadata.id,
                    previous.status.getDescription(),
                )
                // Calling stop() here may be superfluous (depending on the state in which the pipeline is in),
                // but we want to ensure that we clean up the KafkaStreams state of the pipeline because
                // otherwise we have issues in re-starting it.
                // Calling stop() on an already stopped pipeline is safe.
                previous.stop()
            }
        } else { // pipeline doesn't exist
            logger.info(
                "Creating pipeline {pipelineName} version: {pipelineVersion} id: {pipelineId}",
                metadata.name,
                metadata.version,
                metadata.id,
            )
        }

        val (pipeline, err) =
            Pipeline.forSteps(
                metadata,
                steps,
                kafkaProperties,
                kafkaDomainParams,
                kafkaConsumerGroupIdPrefix,
                namespace,
            )
        if (err != null) {
            err.log(logger, Level.ERROR)
            client.pipelineUpdateEvent(
                makePipelineUpdateEvent(
                    metadata = metadata,
                    operation = PipelineOperation.Create,
                    success = false,
                    reason = err.getDescription() ?: "failed to initialize dataflow engine",
                ),
            )
            return
        }

        pipeline!! // assert pipeline is not null when err is null
        if (pipeline.size != steps.size) {
            pipeline.stop()
            client.pipelineUpdateEvent(
                makePipelineUpdateEvent(
                    metadata = metadata,
                    operation = PipelineOperation.Create,
                    success = false,
                    reason = "failed to create all pipeline steps",
                ),
            )

            return
        }

        // This overwrites any previous pipelines with the same id. We can only get here if those previous pipelines
        // are in a failed state and they are being re-created by the scheduler.
        pipelines[metadata.id] = pipeline
        val pipelineStatus: PipelineStatus
        val errTopics = kafkaAdmin.ensureTopicsExist(steps)
        if (errTopics == null) {
            pipelineStatus = pipeline.start()
        } else {
            pipelineStatus =
                PipelineStatus.Error(null)
                    .withException(errTopics)
                    .withMessage("kafka streams topic creation error")
            pipeline.stop()
        }

        // We don't want to mark the PipelineOperation.Create as successful unless the
        // pipeline has started. While states such as "StreamStarting" or "StreamStopped" are
        // not in themselves errors, they are not expected at this stage. If the pipeline
        // is not running here then it can't be marked as ready.
        if (pipelineStatus !is PipelineStatus.Started) {
            pipelineStatus.hasError = true
        }
        pipelineStatus.log(logger, Level.DEBUG)
        client.pipelineUpdateEvent(
            makePipelineUpdateEvent(
                metadata = metadata,
                operation = PipelineOperation.Create,
                success = !pipelineStatus.isError(),
                reason = pipelineStatus.getDescription() ?: defaultReason,
            ),
        )
    }

    private suspend fun handleDelete(metadata: PipelineMetadata) {
        logger.info(
            "Delete pipeline {pipelineName} version: {pipelineVersion} id: {pipelineId}",
            metadata.name,
            metadata.version,
            metadata.id,
        )
        pipelines
            .remove(metadata.id)
            ?.also { pipeline ->
                runBlocking {
                    pipeline.stop()
                }
            }
        client.pipelineUpdateEvent(
            makePipelineUpdateEvent(
                metadata = metadata,
                operation = PipelineOperation.Delete,
                success = true,
                reason = "pipeline removed",
            ),
        )
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

    private fun makePipelineUpdateEvent(
        metadata: PipelineMetadata,
        operation: PipelineOperation,
        success: Boolean,
        reason: String = "",
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
                    .build(),
            )
            .build()
    }

    companion object {
        private val logger = coLogger(PipelineSubscriber::class)
    }
}
