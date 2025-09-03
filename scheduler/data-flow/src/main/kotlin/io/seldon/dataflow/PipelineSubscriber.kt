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
import io.seldon.dataflow.kafka.KafkaAdmin
import io.seldon.dataflow.kafka.KafkaAdminProperties
import io.seldon.dataflow.kafka.KafkaDomainParams
import io.seldon.dataflow.kafka.KafkaProperties
import io.seldon.dataflow.kafka.KafkaStreamsParams
import io.seldon.dataflow.kafka.Pipeline
import io.seldon.dataflow.kafka.PipelineId
import io.seldon.dataflow.kafka.PipelineMetadata
import io.seldon.dataflow.kafka.PipelineTaskFactory
import io.seldon.dataflow.kafka.Task
import io.seldon.dataflow.kafka.TopicWaitRetryParams
import io.seldon.mlops.chainer.ChainerGrpcKt
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineStepUpdate
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineSubscriptionRequest
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateMessage
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateMessage.PipelineOperation
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateStatusMessage
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.FlowPreview
import kotlinx.coroutines.Job
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.asCoroutineDispatcher
import kotlinx.coroutines.async
import kotlinx.coroutines.awaitAll
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.collect
import kotlinx.coroutines.flow.onCompletion
import kotlinx.coroutines.flow.onEach
import kotlinx.coroutines.launch
import kotlinx.coroutines.runBlocking
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
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
    nThreads: Int,
    private val queueCleanupDelayMs: Long = 30_000L,
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
    val dispatcher = Executors.newFixedThreadPool(nThreads).asCoroutineDispatcher()
    val scope = CoroutineScope(SupervisorJob() + dispatcher)

    val pipelines = ConcurrentHashMap<PipelineId, Pipeline>()
    private val queues = ConcurrentHashMap<PipelineId, QueueInfo>()
    private val queuesMutex = Mutex()

    // Task factory for creating pipeline operation tasks
    private val taskFactory =
        PipelineTaskFactory(
            pipelineSubscriber = this,
            kafkaAdmin = kafkaAdmin,
            kafkaProperties = kafkaProperties,
            kafkaDomainParams = kafkaDomainParams,
            name = name,
            logger = logger,
        )

    // Track queues scheduled for deletion
    private data class QueueInfo(
        val queue: Channel<Task>,
        val processingJob: Job,
        var isMarkedForDeletion: Boolean = false,
        var deletionScheduledAt: Long = 0L,
    )

    init {
        // Start background cleanup task
        scope.launch {
            while (true) {
                delay(5000L)
                cleanupMarkedQueues()
            }
        }
    }

    private suspend fun cleanupMarkedQueues() {
        val currentTime = System.currentTimeMillis()
        val toCleanup = mutableListOf<Pair<PipelineId, QueueInfo>>()

        queuesMutex.withLock {
            queues.forEach { (pipelineId, queueInfo) ->
                if (queueInfo.isMarkedForDeletion &&
                    currentTime - queueInfo.deletionScheduledAt > queueCleanupDelayMs
                ) {
                    toCleanup.add(pipelineId to queueInfo)
                }
            }
        }

        toCleanup.forEach { (pipelineId, queueInfo) ->
            queuesMutex.withLock {
                val currentQueueInfo = queues[pipelineId]
                if (currentQueueInfo == queueInfo && currentQueueInfo.isMarkedForDeletion) {
                    logger.debug("Cleaning up queue for pipeline $pipelineId after delay")
                    try {
                        queueInfo.queue.close()
                        queues.remove(pipelineId)
                        logger.debug("Removed pipeline queue from map: $pipelineId")
                    } catch (e: Exception) {
                        logger.error("Error during queue cleanup for pipeline $pipelineId: ${e.message}", e)
                        queues.remove(pipelineId)
                    }
                } else {
                    logger.debug("Queue for pipeline $pipelineId was recreated or unmarked, skipping cleanup")
                }
            }

            try {
                queueInfo.processingJob.join()
            } catch (e: Exception) {
                logger.error("Error waiting for processing job to finish for pipeline $pipelineId: ${e.message}", e)
            }
        }
    }

    private fun startProcessing(
        scope: CoroutineScope,
        queue: Channel<Task>,
    ): Job {
        return scope.launch {
            for (task in queue) {
                try {
                    task.run()
                } catch (e: Exception) {
                    // Useful for debugging - All the failure cases should be already
                    // accounted for in the run method. In the future, we might consider
                    // sending and update back to the scheduler in case the task fails for
                    // some reason.
                    logger.error("Task failed permanently: ${e.message}")
                }
            }
        }
    }

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
        queuesMutex.withLock {
            val existingQueueInfo = queues[metadata.id]
            if (existingQueueInfo == null) {
                // Create new queue and processing job
                val queue = Channel<Task>(Channel.CONFLATED)
                val processingJob = startProcessing(scope, queue)
                queues[metadata.id] = QueueInfo(queue, processingJob)
            } else if (existingQueueInfo.isMarkedForDeletion) {
                // Unmark for deletion since we're recreating
                logger.debug("Unmarking queue for deletion due to recreate request for pipeline ${metadata.id}")
                existingQueueInfo.isMarkedForDeletion = false
                existingQueueInfo.deletionScheduledAt = 0L
            }
        }

        queues[metadata.id]?.queue?.send(
            taskFactory.createTask(
                operation = PipelineOperation.Create,
                metadata = metadata,
                steps = steps,
                timestamp = timestamp,
                kafkaConsumerGroupIdPrefix = kafkaConsumerGroupIdPrefix,
                namespace = namespace,
            )!!,
        )
    }

    private suspend fun handleDelete(
        metadata: PipelineMetadata,
        steps: List<PipelineStepUpdate>,
        timestamp: Long,
    ) {
        var queueToSendTask: Channel<Task>? = null
        queuesMutex.withLock {
            val queueInfo = queues[metadata.id]
            if (queueInfo != null) {
                // Mark for delayed deletion
                logger.debug("Marking queue for delayed deletion for pipeline ${metadata.id}")
                queueInfo.isMarkedForDeletion = true
                queueInfo.deletionScheduledAt = System.currentTimeMillis()
                queueToSendTask = queueInfo.queue
            }
        }

        queueToSendTask?.send(
            taskFactory.createTask(
                operation = PipelineOperation.Delete,
                metadata = metadata,
                steps = steps,
                timestamp = timestamp,
            )!!,
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
