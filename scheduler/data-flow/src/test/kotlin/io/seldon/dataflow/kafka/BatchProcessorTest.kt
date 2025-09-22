/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow.kafka

import io.seldon.mlops.inference.v2.V2Dataplane
import io.seldon.mlops.inference.v2.V2Dataplane.ModelInferRequest
import io.seldon.test.utils.MockFixedKeyProcessorContext
import org.apache.kafka.common.serialization.Serdes
import org.apache.kafka.streams.processor.api.InternalFixedKeyRecordFactory
import org.apache.kafka.streams.processor.api.Record
import org.apache.kafka.streams.state.Stores
import org.junit.jupiter.api.Test
import org.junit.jupiter.params.ParameterizedTest
import org.junit.jupiter.params.provider.Arguments
import org.junit.jupiter.params.provider.Arguments.arguments
import org.junit.jupiter.params.provider.MethodSource
import strikt.api.expectThat
import strikt.assertions.isEmpty
import strikt.assertions.isEqualTo
import strikt.assertions.isGreaterThan
import java.util.stream.Stream

internal class BatchProcessorTest {
    @Test
    fun `should merge requests with standard tensor contents`() {
        val mockContext =
            MockFixedKeyProcessorContext<String, ModelInferRequest>()
        val batcher = BatchProcessor(3)
        val expected = makeRequest("3", listOf(12.34F, 12.34F, 12.34F))
        val requests =
            listOf(
                makeRequest("1", listOf(12.34F)),
                makeRequest("2", listOf(12.34F)),
                makeRequest("3", listOf(12.34F)),
            )

        batcher.init(mockContext)
        val merged = batcher.merge(requests)

        expectThat(merged).isEqualTo(expected)
    }

    @Test
    fun `should merge requests with no tensor contents`() {
        val mockContext =
            MockFixedKeyProcessorContext<String, ModelInferRequest>()
        val batcher = BatchProcessor(3)
        val expected = makeRequest("3", emptyList())
        val requests =
            listOf(
                makeRequest("1", emptyList()),
                makeRequest("2", emptyList()),
                makeRequest("3", emptyList()),
            )

        batcher.init(mockContext)
        val merged = batcher.merge(requests)

        expectThat(merged).isEqualTo(expected)
    }

    @Test
    fun `should merge requests with raw tensor contents`() {
        val mockContext =
            MockFixedKeyProcessorContext<String, ModelInferRequest>()
        val batcher = BatchProcessor(3)
        val expected = makeRequest("3", listOf(12.34F, 12.34F, 12.34F)).withBinaryContents()
        val requests =
            listOf(
                makeRequest("1", listOf(12.34F)).withBinaryContents(),
                makeRequest("2", listOf(12.34F)).withBinaryContents(),
                makeRequest("3", listOf(12.34F)).withBinaryContents(),
            )

        batcher.init(mockContext)
        val merged = batcher.merge(requests)

        expectThat(merged).isEqualTo(expected)
    }

    @Test
    fun `should only forward when batch size met`() {
        val mockContext = MockFixedKeyProcessorContext<String, ModelInferRequest>()
        val store =
            Stores
                .keyValueStoreBuilder(
                    Stores.inMemoryKeyValueStore(BatchProcessor.STATE_STORE_ID),
                    Serdes.String(),
                    Serdes.ByteArray(),
                )
                .withCachingDisabled()
                .withLoggingDisabled()
                .build()
        val batchSize = 10
        val batcher = BatchProcessor(batchSize)
        val streamKey = "789"

        store.init(mockContext.stateStoreContext, store)
        mockContext.setTopic("seldon.foo.model.bar.inputs")
        mockContext.setPartition(3)
        batcher.init(mockContext)

        (1 until batchSize).forEach {
            val preBatchRequest = makeRequest(it.toString(), listOf(it.toFloat()))
            val actual = batcher.process(InternalFixedKeyRecordFactory.create(Record(streamKey, preBatchRequest, it.toLong())))
            expectThat(mockContext.forwarded()).isEmpty()
            expectThat(store.approximateNumEntries()).isGreaterThan(0)
        }

        val batchRequest = makeRequest(batchSize.toString(), listOf(batchSize.toFloat()))
        batcher.process(InternalFixedKeyRecordFactory.create(Record(streamKey, batchRequest, batchSize.toLong())))
        val expected =
            InternalFixedKeyRecordFactory.create(
                Record(
                    streamKey,
                    makeRequest(
                        batchRequest.id,
                        (1..batchSize).map { it.toFloat() },
                    ),
                    batchSize.toLong(),
                ),
            )
        expectThat(mockContext.forwarded().size).isEqualTo(1)
        expectThat(mockContext.forwarded()[0].record()).isEqualTo(expected)
        expectThat(store.approximateNumEntries()).isEqualTo(0)

        // make sure there are no more forwards until twice the batch size
        (1 + batchSize until 2 * batchSize).forEach {
            val postBatchRequest = makeRequest(it.toString(), listOf(it.toFloat()))
            batcher.process(InternalFixedKeyRecordFactory.create(Record(streamKey, postBatchRequest, it.toLong())))
            expectThat(mockContext.forwarded().size).isEqualTo(1)
            expectThat(store.approximateNumEntries()).isGreaterThan(0)
        }
    }

    @ParameterizedTest(name = "{0}")
    @MethodSource("mixedRawAndStandardTensorRequests")
    fun `should support mixed raw and standard tensors`(
        testName: String,
        expected: ModelInferRequest,
        inputs: List<ModelInferRequest>,
    ) {
        val mockContext = MockFixedKeyProcessorContext<String, ModelInferRequest>()
        val batchSize = 3
        val batcher = BatchProcessor(batchSize)

        batcher.init(mockContext)
        val merged = batcher.merge(inputs)

        expectThat(merged).isEqualTo(expected)
    }

    companion object {
        @JvmStatic
        private fun mixedRawAndStandardTensorRequests(): Stream<Arguments> {
            return Stream.of(
                arguments(
                    "standard first",
                    makeRequest("raw", listOf(12.34F, 23.45F)).withBinaryContents(),
                    listOf(
                        makeRequest("standard", listOf(12.34F)),
                        makeRequest("raw", listOf(23.45F)).withBinaryContents(),
                    ),
                ),
                arguments(
                    "raw first",
                    makeRequest("standard", listOf(23.45F, 12.34F)).withBinaryContents(),
                    listOf(
                        makeRequest("raw", listOf(23.45F)).withBinaryContents(),
                        makeRequest("standard", listOf(12.34F)),
                    ),
                ),
                arguments(
                    "raw in the middle",
                    makeRequest("standard", listOf(12.34F, 23.45F, 34.56F)).withBinaryContents(),
                    listOf(
                        makeRequest("standard", listOf(12.34F)),
                        makeRequest("raw", listOf(23.45F)).withBinaryContents(),
                        makeRequest("standard", listOf(34.56F)),
                    ),
                ),
                arguments(
                    "standard in the middle",
                    makeRequest("raw", listOf(23.45F, 34.56F, 45.67F)).withBinaryContents(),
                    listOf(
                        makeRequest("raw", listOf(23.45F)).withBinaryContents(),
                        makeRequest("standard", listOf(34.56F)),
                        makeRequest("raw", listOf(45.67F)).withBinaryContents(),
                    ),
                ),
            )
        }

        private fun makeRequest(
            id: String,
            values: List<Float>,
        ): ModelInferRequest {
            return ModelInferRequest
                .newBuilder()
                .setId(id)
                .addInputs(
                    ModelInferRequest.InferInputTensor
                        .newBuilder()
                        .setName("preprocessed_image")
                        .setDatatype("FP32")
                        .addAllShape(
                            listOf(values.size.toLong(), 1, 1, 1),
                        )
                        .setContents(
                            V2Dataplane.InferTensorContents
                                .newBuilder()
                                .addAllFp32Contents(values)
                                .build(),
                        ),
                )
                .build()
        }
    }
}
