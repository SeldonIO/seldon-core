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
import io.seldon.dataflow.kafka.*
import io.seldon.mlops.chainer.ChainerGrpcKt
import io.seldon.mlops.chainer.ChainerOuterClass.*
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateMessage.PipelineOperation
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
    private val channel = ManagedChannelBuilder
        .forAddress(upstreamHost, upstreamPort)
        .defaultServiceConfig(grpcServiceConfig)
        .usePlaintext() // Use TLS
        .enableRetry()
        .build()
    private val client = ChainerGrpcKt.ChainerCoroutineStub(channel)

    private val pipelines = ConcurrentHashMap<PipelineId, Pipeline>()

    suspend fun subscribe() {
        while (true) {
            logger.info("will connect to ${upstreamHost}:${upstreamPort}")
            retry(binaryExponentialBackoff(50..5_000L)) {
                logger.debug("retrying to connect to ${upstreamHost}:${upstreamPort}")
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
    private suspend fun subscribePipelines(kafkaConsumerGroupIdPrefix: String, namespace: String) {
        logger.info("Subscribing to pipeline updates")
        client
            .subscribePipelineUpdates(request = makeSubscriptionRequest())
            .onEach { update ->
                logger.info("received request for ${update.pipeline}:${update.version} Id:${update.uid}")

                val metadata = PipelineMetadata(
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
                            if(it.value.status.isError) {
                                logger.debug("(bug) pipeline in error state when subscription terminates with error. pipeline id: {pipelineId}", it.key)
                                it.value.stop()
                            }
                        }
                    logger.error("pipeline subscription terminated with error ${cause}")
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
        logger.info(
            "Create pipeline {pipelineName}  version: {pipelineVersion} id: {pipelineId}",
            metadata.name,
            metadata.version,
            metadata.id
        )
        val (pipeline, err) = Pipeline.forSteps(
            metadata,
            steps,
            kafkaProperties,
            kafkaDomainParams,
            kafkaConsumerGroupIdPrefix,
            namespace
        )
        if (err != null) {
            err.log(logger, Level.ERROR)
            client.pipelineUpdateEvent(
                makePipelineUpdateEvent(
                    metadata = metadata,
                    operation = PipelineOperation.Create,
                    success = false,
                    reason = err.getDescription() ?: "failed to initialize dataflow engine"
                )
            )
            return
        }

        pipeline!!  //assert pipeline is not null when err is null
        if (pipeline.size != steps.size) {
            pipeline.stop()
            client.pipelineUpdateEvent(
                makePipelineUpdateEvent(
                    metadata = metadata,
                    operation = PipelineOperation.Create,
                    success = false,
                    reason = "failed to create all pipeline steps"
                )
            )

            return
        }

        val previous = pipelines.putIfAbsent(metadata.id, pipeline)
        var pipelineStatus: PipelineStatus
        if (previous == null) {
            val err = kafkaAdmin.ensureTopicsExist(steps)
            if (err == null) {
                pipelineStatus = pipeline.start()
            } else {
                pipelineStatus = PipelineStatus.Error(null)
                    .withException(err)
                    .withMessage("kafka streams topic creation error")
                pipeline.stop()
            }
        } else {
            pipelineStatus = previous.status
            logger.warn("pipeline {pipelineName} with id {pipelineId} already exists", metadata.name, metadata.id)
            if (pipelineStatus.isError) {
                // do not try to resuscitate an existing pipeline if in a failed state
                // it's up to the scheduler to delete it & reinitialize it, as it might require
                // coordination with {model, pipeline}gateway
                previous.stop()
            }
        }

        // There is a small chance that pipeline.start() returned a status of PipelineState.StreamStopped(),
        // if the process is being signalled to shutdown during its execution, and calls pipeline.stop()
        //
        // For this case, we don't want to mark the Create operation as successful, so we force the state
        // to be an error (despite no actual error having occurred) before sending the update to the scheduler.
        if(pipelineStatus is PipelineStatus.StreamStopped) {
            pipelineStatus.isError = true
        }
        pipelineStatus.log(logger, Level.INFO)
        client.pipelineUpdateEvent(
            makePipelineUpdateEvent(
                metadata = metadata,
                operation = PipelineOperation.Create,
                success = !pipelineStatus.isError,
                reason = pipelineStatus.getDescription() ?: "pipeline created"
            )
        )
    }

    private suspend fun handleDelete(metadata: PipelineMetadata) {
        logger.info("Delete pipeline {pipelineName} version: {pipelineVersion} id: {pipelineId}", metadata.name, metadata.version, metadata.id )
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
            )
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
                    .build()
            )
            .build()
    }

    companion object {
        private val logger = coLogger(PipelineSubscriber::class)
    }
}