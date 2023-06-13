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
import io.seldon.mlops.inference.v2.V2Dataplane.InferParameter
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

    override fun transform(key: String, value: ModelInferRequest): TBatchRequest {
        val stateStoreKey = "random-batch-id"
        var batchedRequest = value
        val reqBytes = aggregateStore.putIfAbsent(stateStoreKey, value.toByteArray())
        if (reqBytes != null) {
            batchedRequest = merge(listOf(ModelInferRequest.parseFrom(reqBytes), value))
            aggregateStore.put(stateStoreKey, batchedRequest.toByteArray())
        }
        val batchSize = batchedRequest.getInputs(0).getShape(0)
        val returnValue = when {
            batchSize >= threshold -> {
                aggregateStore.delete(stateStoreKey)
                KeyValue.pair(key, batchedRequest)
            }
            else -> null
        }
        return returnValue
    }

    private fun merge(requests: List<ModelInferRequest>): ModelInferRequest {
        val tensorNames = requests
            .flatMap { it.inputsList }
            .groupBy { it.name }
            .mapValues { (k, v) ->
                val batchSize = v.sumOf { it.getShape(0) }
                val dataShape = v.first().shapeList.drop(1)
                val shape = mutableListOf(batchSize) + dataShape
                val parameters = mutableMapOf<String, InferParameter>()
                v.map { it.parametersMap }.forEach {parameters.putAll(it)}
                InferInputTensor
                    .newBuilder()
                    .setName(k)
                    .addAllShape(shape)
                    .putAllParameters(parameters)
                    .setContents(
                        InferTensorContents
                            .newBuilder()
                            .addAllIntContents(v.flatMap { it.contents.intContentsList })
                            .addAllInt64Contents(v.flatMap { it.contents.int64ContentsList })
                            .addAllFp32Contents(v.flatMap { it.contents.fp32ContentsList })
                            .addAllFp64Contents(v.flatMap { it.contents.fp64ContentsList })
                            .addAllBoolContents(v.flatMap { it.contents.boolContentsList })
                            .addAllBytesContents(v.flatMap { it.contents.bytesContentsList })
                            .addAllUintContents(v.flatMap { it.contents.uintContentsList })
                            .addAllUint64Contents(v.flatMap { it.contents.uint64ContentsList })
                            .build()
                    )
                    .setDatatype(v.first().datatype)
                    .build()
            }
        val batchReferenceRequest = requests.last()
        return ModelInferRequest
            .newBuilder()
            .setId(batchReferenceRequest.id)
            .setModelName(batchReferenceRequest.modelName)
            .setModelVersion(batchReferenceRequest.modelVersion)
            .putAllParameters(batchReferenceRequest.parametersMap)
            .addAllInputs(tensorNames.values)
            .build()
    }

    override fun close() {
    }

    companion object {
        private val logger = noCoLogger(BatchProcessor::class)
        const val STATE_STORE_ID = "batch-store"
        val stateStoreBuilder: StoreBuilder<KeyValueStore<String, ByteArray>> = Stores
            .keyValueStoreBuilder(
                Stores.inMemoryKeyValueStore(STATE_STORE_ID),
                Serdes.String(),
                Serdes.ByteArray()
            )
            .withLoggingDisabled()
            .withCachingDisabled()
    }
}