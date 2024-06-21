/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow.kafka

import com.google.protobuf.ByteString
import io.klogging.noCoLogger
import io.seldon.mlops.inference.v2.V2Dataplane.InferTensorContents
import io.seldon.mlops.inference.v2.V2Dataplane.ModelInferRequest
import io.seldon.mlops.inference.v2.V2Dataplane.ModelInferRequest.InferInputTensor
import org.apache.kafka.common.serialization.Serdes
import org.apache.kafka.streams.KeyValue
import org.apache.kafka.streams.kstream.Transformer
import org.apache.kafka.streams.processor.ProcessorContext
import org.apache.kafka.streams.state.KeyValueStore
import org.apache.kafka.streams.state.StoreBuilder
import org.apache.kafka.streams.state.Stores

typealias TBatchRequest = KeyValue<String, ModelInferRequest>?
typealias TBatchStore = KeyValueStore<String, ByteArray>

class BatchProcessor(private val threshold: Int) : Transformer<String, ModelInferRequest, TBatchRequest> {
    private var ctx: ProcessorContext? = null
    private val aggregateStore: TBatchStore by lazy {
        when (ctx) {
            null -> throw IllegalStateException("processor context is null")
            else -> ctx!!.getStateStore(STATE_STORE_ID) as TBatchStore
        }
    }

    override fun init(context: ProcessorContext) {
        this.ctx = context
    }

    override fun transform(
        key: String,
        value: ModelInferRequest,
    ): TBatchRequest {
        val stateStoreKey = "random-batch-id"
        var batchedRequest = value
        val reqBytes = aggregateStore.putIfAbsent(stateStoreKey, value.toByteArray())
        if (reqBytes != null) {
            batchedRequest = merge(listOf(ModelInferRequest.parseFrom(reqBytes), value))
            aggregateStore.put(stateStoreKey, batchedRequest.toByteArray())
        }
        val batchSize = batchedRequest.getInputs(0).getShape(0)
        val returnValue =
            when {
                batchSize >= threshold -> {
                    aggregateStore.delete(stateStoreKey)
                    KeyValue.pair(key, batchedRequest)
                }
                else -> null
            }
        return returnValue
    }

    // merge accepts a list of inference requests and combines them into a single request.
    // Combining broadly means concatenating along tensors and updating their batch sizes.
    // It is assumed all requests have the same tensor data types and shapes (bar batch size).
    // To support joins, the request ID of the combined request is that of the last in the batch.
    // For this reason, batches should be ordered chronologically, with the most recent item last.
    internal fun merge(requests: List<ModelInferRequest>): ModelInferRequest {
        val batchReferenceRequest = requests.last()
        val combinedRequest =
            ModelInferRequest
                .newBuilder()
                .setId(batchReferenceRequest.id)
                .setModelName(batchReferenceRequest.modelName)
                .setModelVersion(batchReferenceRequest.modelVersion)
                .putAllParameters(batchReferenceRequest.parametersMap)

        when {
            requests.any { it.rawInputContentsCount > 0 } -> {
                val (tensors, rawContents) = requests.withBinaryContents().mergeRawTensors()
                combinedRequest.addAllInputs(tensors)
                combinedRequest.addAllRawInputContents(rawContents)
            }
            else -> {
                val tensors = requests.mergeStandardTensors()
                combinedRequest.addAllInputs(tensors)
            }
        }

        return combinedRequest.build()
    }

    private fun List<ModelInferRequest>.mergeStandardTensors(): List<InferInputTensor> {
        return this
            .flatMap { it.inputsList }
            .groupBy { it.name }
            .map { (name, tensors) ->
                val batchSize = tensors.sumOf { it.getShape(0) }
                val elementShape = tensors.first().shapeList.toList().drop(1)
                val newShape = listOf(batchSize) + elementShape

                val parameters =
                    tensors
                        .flatMap { it.parametersMap.entries }
                        .associate { it.toPair() }

                val datatype = tensors.first().datatype

                val contents =
                    InferTensorContents
                        .newBuilder()
                        .apply {
                            when (DataType.valueOf(datatype)) {
                                DataType.UINT8,
                                DataType.UINT16,
                                DataType.UINT32,
                                ->
                                    this.addAllUintContents(
                                        tensors.flatMap { it.contents.uintContentsList },
                                    )
                                DataType.UINT64 ->
                                    this.addAllUint64Contents(
                                        tensors.flatMap { it.contents.uint64ContentsList },
                                    )
                                DataType.INT8,
                                DataType.INT16,
                                DataType.INT32,
                                ->
                                    this.addAllIntContents(
                                        tensors.flatMap { it.contents.intContentsList },
                                    )
                                DataType.INT64 ->
                                    this.addAllInt64Contents(
                                        tensors.flatMap { it.contents.int64ContentsList },
                                    )
                                DataType.FP16, // may need to handle this separately in future
                                DataType.FP32,
                                ->
                                    this.addAllFp32Contents(
                                        tensors.flatMap { it.contents.fp32ContentsList },
                                    )
                                DataType.FP64 ->
                                    this.addAllFp64Contents(
                                        tensors.flatMap { it.contents.fp64ContentsList },
                                    )
                                DataType.BOOL ->
                                    this.addAllBoolContents(
                                        tensors.flatMap { it.contents.boolContentsList },
                                    )
                                DataType.BYTES ->
                                    this.addAllBytesContents(
                                        tensors.flatMap { it.contents.bytesContentsList },
                                    )
                            }
                        }
                        .build()

                InferInputTensor
                    .newBuilder()
                    .setName(name)
                    .setDatatype(datatype)
                    .addAllShape(newShape)
                    .putAllParameters(parameters)
                    .setContents(contents)
                    .build()
            }
    }

    private fun List<ModelInferRequest>.mergeRawTensors(): Pair<List<InferInputTensor>, List<ByteString>> {
        return this
            .flatMap {
                it.rawInputContentsList.zip(it.inputsList)
            }
            .groupBy { it.second.name }
            .map { (name, rawContentsAndMetadata) ->
                val (rawContents, metadata) = rawContentsAndMetadata.unzip()
                val datatype = metadata.first().datatype

                val batchSize = metadata.sumOf { it.getShape(0) }
                val elementShape = metadata.first().shapeList.drop(1)
                val shape = listOf(batchSize) + elementShape

                val parameters = metadata.flatMap { it.parametersMap.entries }.associate { it.toPair() }

                val combinedRawContents = ByteString.copyFrom(rawContents)

                val tensor =
                    InferInputTensor
                        .newBuilder()
                        .setName(name)
                        .setDatatype(datatype)
                        .addAllShape(shape)
                        .putAllParameters(parameters)
                        .build()

                tensor to combinedRawContents
            }
            .unzip()
    }

    override fun close() {
    }

    companion object {
        private val logger = noCoLogger(BatchProcessor::class)
        const val STATE_STORE_ID = "batch-store"
        val stateStoreBuilder: StoreBuilder<KeyValueStore<String, ByteArray>> =
            Stores
                .keyValueStoreBuilder(
                    Stores.inMemoryKeyValueStore(STATE_STORE_ID),
                    Serdes.String(),
                    Serdes.ByteArray(),
                )
                .withLoggingDisabled()
                .withCachingDisabled()
    }
}
