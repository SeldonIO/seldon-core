package io.seldon.dataflow

import com.github.michaelbull.retry.ContinueRetrying
import com.github.michaelbull.retry.StopRetrying
import com.github.michaelbull.retry.policy.RetryPolicy
import com.github.michaelbull.retry.policy.binaryExponentialBackoff
import com.github.michaelbull.retry.policy.plus
import com.github.michaelbull.retry.retry
import io.grpc.ManagedChannelBuilder
import io.grpc.StatusException
import io.grpc.StatusRuntimeException
import io.seldon.dataflow.kafka.*
import io.seldon.mlops.chainer.ChainerGrpcKt
import io.seldon.mlops.chainer.ChainerOuterClass.*
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineStepUpdate.PipelineJoinType
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineUpdateMessage.PipelineOperation
import kotlinx.coroutines.FlowPreview
import kotlinx.coroutines.flow.asFlow
import kotlinx.coroutines.flow.collect
import kotlinx.coroutines.flow.onCompletion
import kotlinx.coroutines.flow.onEach
import kotlinx.coroutines.runBlocking
import java.util.concurrent.ConcurrentHashMap
import io.klogging.logger as coLogger

typealias PipelineId = String

data class PipelineMetadata(
    val id: PipelineId,
    val name: String,
    val version: Int,
)

data class PipelineTopology(
    val metadata: PipelineMetadata,
    val transformers: List<Transformer>,
)

@OptIn(FlowPreview::class)
class PipelineSubscriber(
    private val name: String,
    private val kafkaProperties: KafkaProperties,
    private val kafkaDomainParams: KafkaDomainParams,
    upstreamHost: String,
    upstreamPort: Int,
    grpcServiceConfig: Map<String, Any>,
    ) {
    private val kafkaAdmin = KafkaAdmin(kafkaProperties)
    private val upstreamHost = upstreamHost
    private val upstreamPort = upstreamPort
    private val channel = ManagedChannelBuilder
        .forAddress(upstreamHost, upstreamPort)
        .defaultServiceConfig(grpcServiceConfig)
        .usePlaintext() // Use TLS
        .enableRetry()
        .build()
    private val client = ChainerGrpcKt.ChainerCoroutineStub(channel)

    private val pipelines = ConcurrentHashMap<PipelineId, PipelineTopology>()
    private val grpcFailurePolicy: RetryPolicy<Throwable> = {
        when (reason) {
            is StatusException,
            is StatusRuntimeException -> ContinueRetrying
            else -> StopRetrying
            // TODO - be more intelligent about non-retryable errors (e.g. not implemented)
        }
    }

    suspend fun subscribe() {
        logger.info("Will connect to ${upstreamHost}:${upstreamPort}")
        retry(grpcFailurePolicy + binaryExponentialBackoff(50..5_000L)) {
            subscribePipelines()
        }
    }

    // TODO
    //  - If a topology encounters an error, we should signal back to the scheduler about this.
    //  - If the scheduler updates/removes a topology, we need to cancel the corresponding coroutine.
    //  ...
    //  Pipeline UID should be enough to uniquely key it, even across versions?
    //  ...
    //  - Add map of model name -> (weak) referrents/reference count to avoid recreation of streams
    private suspend fun subscribePipelines() {

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
                    PipelineOperation.Create -> handleCreate(metadata, update.updatesList)
                    PipelineOperation.Delete -> handleDelete(metadata)
                    else -> logger.warn("unrecognised pipeline operation (${update.op})")
                }
            }
            .onCompletion {
                logger.info("pipeline subscription terminated")
            }
            .collect()
        // TODO - error handling?
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
    ) {
        kafkaAdmin.ensureTopicsExist(steps)

        val transformers = steps
            .mapNotNull {
                transformerFor(
                    metadata.name,
                    it.sourcesList,
                    it.triggersList,
                    it.tensorMapMap,
                    it.sink,
                    it.inputJoinTy,
                    it.triggersJoinTy,
                    it.batch,
                    kafkaProperties,
                    kafkaDomainParams,
                )
            }
            .also { transformers ->
                if (transformers.size != steps.size) {
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
            }

        val previous = pipelines
            .putIfAbsent(
                metadata.id,
                PipelineTopology(metadata, transformers),
            )
        if (previous == null) {
            transformers.forEach { it.start() }
        } else {
            logger.warn("pipeline ${metadata.id} already exists")
        }
        client.pipelineUpdateEvent(
            makePipelineUpdateEvent(
                metadata = metadata,
                operation = PipelineOperation.Create,
                success = true,
                reason = "Created pipeline"
            )
        )
    }

    private suspend fun handleDelete(metadata: PipelineMetadata) {
        logger.info("Delete pipeline ${metadata.name}")
        pipelines
            .remove(metadata.id)
            ?.also { pipeline ->
                runBlocking {
                    pipeline.transformers
                        .asFlow()
                        .parallel(
                            scope = this,
                            concurrency = pipeline.transformers.size,
                        ) { step ->
                            cancelPipelineStep(pipeline.metadata, step, "removal requested")
                        }
                        .collect()
                }
            }
        client.pipelineUpdateEvent(
            makePipelineUpdateEvent(
                metadata = metadata,
                operation = PipelineOperation.Delete,
                success = true,
                reason = "Pipeline removed",
            )
        )
    }

    fun cancelPipelines(reason: String) {
        runBlocking {
            logger.info("cancelling pipelines")
            pipelines.values
                .flatMap { pipeline ->
                    pipeline.transformers.map { pipeline.metadata to it }
                }
                .asFlow()
                .parallel(scope = this, concurrency = pipelines.size) { (metadata, transformer) ->
                    cancelPipelineStep(metadata, transformer, reason)
                }
                .collect()
        }
    }

    private suspend fun cancelPipelineStep(
        metadata: PipelineMetadata,
        transformer: Transformer,
        reason: String,
    ) {
        transformer.stop()
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
