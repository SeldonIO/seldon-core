/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow.kafka

import io.seldon.mlops.chainer.ChainerOuterClass
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineTopic
import org.apache.kafka.streams.StreamsBuilder
import org.junit.jupiter.params.ParameterizedTest
import org.junit.jupiter.params.provider.Arguments
import org.junit.jupiter.params.provider.Arguments.arguments
import org.junit.jupiter.params.provider.MethodSource
import strikt.api.Assertion
import strikt.api.expect
import strikt.assertions.*
import java.util.stream.Stream

internal class PipelineStepTest {
    @ParameterizedTest(name = "{0}")
    @MethodSource
    fun stepFor(
        testName: String,
        expected: PipelineStep?,
        sources: List<PipelineTopic>,
    ) {
        val result =
            stepFor(
                StreamsBuilder(),
                defaultPipelineName,
                sources,
                emptyList(),
                emptyList(),
                defaultPipelineTopic,
                ChainerOuterClass.PipelineStepUpdate.PipelineJoinType.Inner,
                ChainerOuterClass.PipelineStepUpdate.PipelineJoinType.Inner,
                ChainerOuterClass.Batch.getDefaultInstance(),
                kafkaDomainParams,
            )

        expect {
            when (expected) {
                null -> that(result).isNull()
                else -> that(result).isNotNull().isSameTypeAs(expected).matches(expected)
            }
        }
    }

    companion object {
        private const val defaultPipelineName = "some-pipeline"
        private val defaultPipelineTopic =
            PipelineTopic.newBuilder()
                .setTopicName("seldon.namespace.sinkModel.inputs")
                .setPipelineName(defaultPipelineName).build()
        private val defaultSink = TopicForPipeline(topicName = "seldon.namespace.sinkModel.inputs", pipelineName = defaultPipelineName)
        private val kafkaDomainParams = KafkaDomainParams(useCleanState = true, joinWindowMillis = 1_000L)

        @JvmStatic
        fun stepFor(): Stream<Arguments> =
            Stream.of(
                arguments("no sources", null, emptyList<PipelineTopic>()),
                arguments(
                    "single source, no tensors",
                    makeChainerFor(
                        inputTopic =
                            TopicForPipeline(
                                topicName = "seldon.namespace.model.model11.outputs",
                                pipelineName = defaultPipelineName,
                            ),
                        tensors = null,
                    ),
                    listOf(
                        PipelineTopic.newBuilder().setTopicName(
                            "seldon.namespace.model.model11.outputs",
                        ).setPipelineName(defaultPipelineName).build(),
                    ),
                ),
                arguments(
                    "single source, one tensor",
                    makeChainerFor(
                        inputTopic =
                            TopicForPipeline(
                                topicName = "seldon.namespace.model.model1.outputs",
                                pipelineName = defaultPipelineName,
                            ),
                        tensors = setOf("tensorA"),
                    ),
                    listOf(
                        PipelineTopic.newBuilder().setTopicName(
                            "seldon.namespace.model.model1.outputs",
                        ).setPipelineName(defaultPipelineName).setTensor("tensorA").build(),
                    ),
                ),
                arguments(
                    "single source, multiple tensors",
                    makeChainerFor(
                        inputTopic =
                            TopicForPipeline(
                                topicName = "seldon.namespace.model.model1.outputs",
                                pipelineName = defaultPipelineName,
                            ),
                        tensors = setOf("tensorA", "tensorB"),
                    ),
                    listOf(
                        PipelineTopic.newBuilder().setTopicName("seldon.namespace.model.model1.outputs").setPipelineName(
                            defaultPipelineName,
                        ).setTensor("tensorA").build(),
                        PipelineTopic.newBuilder().setTopicName("seldon.namespace.model.model1.outputs").setPipelineName(
                            defaultPipelineName,
                        ).setTensor("tensorB").build(),
                    ),
                ),
                arguments(
                    "multiple sources, no tensors",
                    makeJoinerFor(
                        inputTopics =
                            setOf(
                                TopicForPipeline(topicName = "seldon.namespace.model.modelA.outputs", pipelineName = defaultPipelineName),
                                TopicForPipeline(topicName = "seldon.namespace.model.modelB.outputs", pipelineName = defaultPipelineName),
                            ),
                        tensorsByTopic = null,
                    ),
                    listOf(
                        PipelineTopic.newBuilder().setTopicName("seldon.namespace.model.modelA.outputs").setPipelineName(
                            defaultPipelineName,
                        ).build(),
                        PipelineTopic.newBuilder().setTopicName("seldon.namespace.model.modelB.outputs").setPipelineName(
                            defaultPipelineName,
                        ).build(),
                    ),
                ),
                arguments(
                    "multiple sources, multiple tensors",
                    makeJoinerFor(
                        inputTopics =
                            setOf(
                                TopicForPipeline(topicName = "seldon.namespace.model.modelA.outputs", pipelineName = defaultPipelineName),
                                TopicForPipeline(topicName = "seldon.namespace.model.modelB.outputs", pipelineName = defaultPipelineName),
                            ),
                        tensorsByTopic =
                            mapOf(
                                TopicForPipeline(topicName = "seldon.namespace.model.modelA.outputs", pipelineName = defaultPipelineName) to setOf("tensor1"),
                                TopicForPipeline(topicName = "seldon.namespace.model.modelB.outputs", pipelineName = defaultPipelineName) to setOf("tensor2"),
                            ),
                    ),
                    listOf(
                        PipelineTopic.newBuilder().setTopicName("seldon.namespace.model.modelA.outputs").setPipelineName(
                            defaultPipelineName,
                        ).setTensor("tensor1").build(),
                        PipelineTopic.newBuilder().setTopicName("seldon.namespace.model.modelB.outputs").setPipelineName(
                            defaultPipelineName,
                        ).setTensor("tensor2").build(),
                    ),
                ),
                arguments(
                    "tensors override plain topic",
                    makeChainerFor(
                        inputTopic =
                            TopicForPipeline(
                                topicName = "seldon.namespace.model.modelA.outputs",
                                pipelineName = defaultPipelineName,
                            ),
                        tensors = setOf("tensorA"),
                    ),
                    listOf(
                        PipelineTopic.newBuilder().setTopicName("seldon.namespace.model.modelA.outputs").setPipelineName(
                            defaultPipelineName,
                        ).setTensor("tensorA").build(),
                        PipelineTopic.newBuilder().setTopicName("seldon.namespace.model.modelA.outputs").setPipelineName(
                            defaultPipelineName,
                        ).build(),
                    ),
                ),
            )

        private fun makeChainerFor(
            inputTopic: TopicForPipeline,
            tensors: Set<TensorName>?,
        ): Chainer =
            Chainer(
                StreamsBuilder(),
                inputTopic = inputTopic,
                tensors = tensors,
                pipelineName = defaultPipelineName,
                outputTopic = defaultSink,
                tensorRenaming = emptyList(),
                kafkaDomainParams = kafkaDomainParams,
                inputTriggerTopics = emptySet(),
                triggerJoinType = ChainerOuterClass.PipelineStepUpdate.PipelineJoinType.Inner,
                triggerTensorsByTopic = emptyMap(),
                batchProperties = ChainerOuterClass.Batch.getDefaultInstance(),
            )

        private fun makeJoinerFor(
            inputTopics: Set<TopicForPipeline>,
            tensorsByTopic: Map<TopicForPipeline, Set<TensorName>>?,
        ): Joiner =
            Joiner(
                StreamsBuilder(),
                inputTopics = inputTopics,
                tensorsByTopic = tensorsByTopic,
                pipelineName = defaultPipelineName,
                outputTopic = defaultSink,
                tensorRenaming = emptyList(),
                kafkaDomainParams = kafkaDomainParams,
                joinType = ChainerOuterClass.PipelineStepUpdate.PipelineJoinType.Inner,
                inputTriggerTopics = emptySet(),
                triggerJoinType = ChainerOuterClass.PipelineStepUpdate.PipelineJoinType.Inner,
                triggerTensorsByTopic = emptyMap(),
            )
    }
}

fun Assertion.Builder<PipelineStep>.isSameTypeAs(other: PipelineStep) =
    assert("Same type") {
        when {
            it::class == other::class -> pass()
            else -> fail(actual = other::class.simpleName)
        }
    }

fun Assertion.Builder<PipelineStep>.matches(expected: PipelineStep) =
    assert("Type and values are the same") {
        when {
            it is Chainer && expected is Chainer ->
                expect {
                    that(it) {
                        get { inputTopic }.isEqualTo(expected.inputTopic)
                        get { outputTopic }.isEqualTo(expected.outputTopic)
                        get { tensors }.isEqualTo(expected.tensors)
                    }
                }
            it is Joiner && expected is Joiner ->
                expect {
                    that(it) {
                        get { inputTopics }.isEqualTo(expected.inputTopics)
                        get { outputTopic }.isEqualTo(expected.outputTopic)
                        get { tensorsByTopic }.isEqualTo(expected.tensorsByTopic)
                        get { tensorRenaming }.isEqualTo(expected.tensorRenaming)
                        get { kafkaDomainParams }.isEqualTo(expected.kafkaDomainParams)
                    }
                }
            else -> fail(actual = expected)
        }
    }
