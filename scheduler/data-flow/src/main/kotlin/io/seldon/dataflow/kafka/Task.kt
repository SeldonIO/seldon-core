package io.seldon.dataflow.kafka

import io.klogging.Klogger
import io.klogging.Level
import io.seldon.dataflow.PipelineSubscriber
import io.seldon.dataflow.withException
import io.seldon.dataflow.withMessage
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineStepUpdate
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateMessage.PipelineOperation
import kotlinx.coroutines.runBlocking
import kotlin.collections.set

abstract class Task {
    abstract suspend fun run()
}

class CreationTask(
    private val pipelineSubscriber: PipelineSubscriber,
    private val metadata: PipelineMetadata,
    private val steps: List<PipelineStepUpdate>,
    private val kafkaAdmin: KafkaAdmin,
    private val kafkaProperties: KafkaProperties,
    private val kafkaDomainParams: KafkaDomainParams,
    private val kafkaConsumerGroupIdPrefix: String,
    private val namespace: String,
    private val timestamp: Long,
    private val name: String,
    private val logger: Klogger,
) : Task() {
    override suspend fun run() {
        logger.info(
            "Creating pipeline {pipelineName} version: {pipelineVersion} id: {pipelineId}",
            metadata.name,
            metadata.version,
            metadata.id,
        )

        val pipeline = pipelineSubscriber.pipelines[metadata.id]
        val err =
            pipeline?.forSteps(
                metadata,
                steps,
                kafkaProperties,
                kafkaDomainParams,
                kafkaConsumerGroupIdPrefix,
                namespace,
            )
        if (err != null) {
            err.log(logger, Level.ERROR)
            pipelineSubscriber.client.pipelineUpdateEvent(
                pipelineSubscriber.makePipelineUpdateEvent(
                    metadata = metadata,
                    operation = PipelineOperation.Create,
                    success = false,
                    reason = err.getDescription() ?: "failed to initialize dataflow engine",
                    timestamp = timestamp,
                    stream = name,
                ),
            )
            return
        }

        pipeline!! // assert pipeline is not null when err is null
        if (pipeline.size != steps.size) {
            pipeline.stop()
            pipelineSubscriber.client.pipelineUpdateEvent(
                pipelineSubscriber.makePipelineUpdateEvent(
                    metadata = metadata,
                    operation = PipelineOperation.Create,
                    success = false,
                    reason = "failed to create all pipeline steps",
                    timestamp = timestamp,
                    stream = name,
                ),
            )
            return
        }

        // This overwrites any previous pipelines with the same id. We can only get here if those previous pipelines
        // are in a failed state, and they are being re-created by the scheduler.
        pipelineSubscriber.pipelines[metadata.id] = pipeline
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
        val defaultReason = "pipeline created"

        pipelineSubscriber.client.pipelineUpdateEvent(
            pipelineSubscriber.makePipelineUpdateEvent(
                metadata = metadata,
                operation = PipelineOperation.Create,
                success = !pipelineStatus.isError(),
                reason = pipelineStatus.getDescription() ?: defaultReason,
                timestamp = timestamp,
                stream = name,
            ),
        )
    }
}

class DeletionTask(
    private val pipeline: Pipeline,
    private val pipelineSubscriber: PipelineSubscriber,
    private val metadata: PipelineMetadata,
    private val steps: List<PipelineStepUpdate>,
    private val kafkaAdmin: KafkaAdmin,
    private val timestamp: Long,
    private val name: String,
    private val logger: Klogger,
) : Task() {
    override suspend fun run() {
        logger.info(
            "Delete pipeline {pipelineName} version: {pipelineVersion} id: {pipelineId}",
            metadata.name,
            metadata.version,
            metadata.id,
        )

        // stop kafka stream
        runBlocking {
            pipeline.stop()
        }

        var pipelineError: PipelineStatus? = null
        val errTopics = kafkaAdmin.deleteTopics(steps)
        if (errTopics != null) {
            pipelineError =
                PipelineStatus.Error(null)
                    .withException(errTopics)
                    .withMessage("kafka streams topic deletion error")
        }

        pipelineSubscriber.client.pipelineUpdateEvent(
            pipelineSubscriber.makePipelineUpdateEvent(
                metadata = metadata,
                operation = PipelineOperation.Delete,
                success = pipelineError == null,
                reason = pipelineError?.getDescription() ?: "pipeline removed",
                timestamp = timestamp,
                stream = name,
            ),
        )
    }
}

class UpdateTask(
    private val pipelineSubscriber: PipelineSubscriber,
    private val metadata: PipelineMetadata,
    private val timestamp: Long,
    private val name: String,
    private val logger: Klogger,
) : Task() {
    override suspend fun run() {
        val defaultReason = "pipeline created"
        val previous = pipelineSubscriber.pipelines[metadata.id]!!

        if (previous.status.isActive()) {
            pipelineSubscriber.client.pipelineUpdateEvent(
                pipelineSubscriber.makePipelineUpdateEvent(
                    metadata = metadata,
                    operation = PipelineOperation.Create,
                    success = true,
                    reason = previous.status.getDescription() ?: defaultReason,
                    timestamp = timestamp,
                    stream = name,
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
        }

        // pipeline exists but in failed/stopped state; cleanup state and re-create
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
}

class StopTask(
    private val pipelineSubscriber: PipelineSubscriber,
    private val metadata: PipelineMetadata,
    private val logger: Klogger,
) : Task() {
    override suspend fun run() {
        val previous = pipelineSubscriber.pipelines[metadata.id]!!

        // pipeline exists but in failed/stopped state; cleanup state and re-create
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
}
