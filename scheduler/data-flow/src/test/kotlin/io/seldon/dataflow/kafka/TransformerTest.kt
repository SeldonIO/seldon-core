package io.seldon.dataflow.kafka

import io.seldon.mlops.chainer.ChainerOuterClass
import org.junit.jupiter.params.ParameterizedTest
import org.junit.jupiter.params.provider.Arguments
import org.junit.jupiter.params.provider.Arguments.arguments
import org.junit.jupiter.params.provider.MethodSource
import strikt.api.Assertion
import strikt.api.expect
import strikt.api.expectThat
import strikt.assertions.*
import java.util.stream.Stream

internal class TransformerTest {

    @ParameterizedTest(name = "{0}")
    @MethodSource
    fun areTensorsFromSameTopic(testName: String, expected: Boolean, sources: List<TopicName>) {
        val (actual, _) = sources.areTensorsFromSameTopic()
        expectThat(expected).isEqualTo(actual)
    }

    @ParameterizedTest(name = "{0}")
    @MethodSource
    fun transformerFor(
        testName: String,
        expected: Transformer?,
        sources: List<TopicName>,
    ) {
        val result =
            transformerFor(
                defaultPipelineName,
                sources,
                emptyList(),
                emptyMap(),
                defaultSink,
                ChainerOuterClass.PipelineStepUpdate.PipelineJoinType.Inner,
                ChainerOuterClass.PipelineStepUpdate.PipelineJoinType.Inner,
                baseKafkaProperties,
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
        private const val defaultSink = "seldon.namespace.sinkModel.inputs"
        private const val defaultPipelineName = "some-pipeline"
        private val baseKafkaProperties = getKafkaProperties(
            KafkaStreamsParams(bootstrapServers = "", numCores = 0),
        )
        private val kafkaDomainParams = KafkaDomainParams(useCleanState = true, joinWindowMillis = 1_000L)

        @JvmStatic
        fun areTensorsFromSameTopic(): Stream<Arguments> =
            Stream.of(
                arguments("no sources", false, emptyList<String>()),
                arguments("no tensors", false, listOf("seldon.namespace.model.outputs")),
                arguments("single tensor", true, listOf("seldon.namespace.model.outputs.tensorA")),
                arguments(
                    "tensors from same model",
                    true,
                    listOf(
                        "seldon.namespace.model.outputs.tensorA",
                        "seldon.namespace.model.outputs.tensorB",
                    ),
                ),
                arguments(
                    "tensors from different models",
                    false,
                    listOf(
                        "seldon.namespace.model1.outputs.tensorA",
                        "seldon.namespace.model2.outputs.tensorA",
                    ),
                ),
            )

        @JvmStatic
        fun transformerFor(): Stream<Arguments> =
            Stream.of(
                arguments("no sources", null, emptyList<String>()),
                arguments(
                    "single source, no tensors",
                    makeChainerFor(
                        inputTopic = "seldon.namespace.model.outputs",
                        tensors = null,
                    ),
                    listOf("seldon.namespace.model.outputs"),
                ),
                arguments(
                    "single source, one tensor",
                    makeChainerFor(
                        inputTopic = "seldon.namespace.model.outputs",
                        tensors = setOf("tensorA")
                    ),
                    listOf("seldon.namespace.model.outputs.tensorA"),
                ),
                arguments(
                    "single source, multiple tensors",
                    makeChainerFor(
                        inputTopic = "seldon.namespace.model.outputs",
                        tensors = setOf("tensorA", "tensorB")
                    ),
                    listOf(
                        "seldon.namespace.model.outputs.tensorA",
                        "seldon.namespace.model.outputs.tensorB",
                    ),
                ),
                arguments(
                    "multiple sources, no tensors",
                    makeJoinerFor(
                        inputTopics = setOf("seldon.namespace.modelA.outputs", "seldon.namespace.modelB.outputs"),
                        tensorsByTopic = null,
                    ),
                    listOf(
                        "seldon.namespace.modelA.outputs",
                        "seldon.namespace.modelB.outputs",
                    ),
                ),
                arguments(
                    "multiple sources, multiple tensors",
                    makeJoinerFor(
                        inputTopics = setOf(
                            "seldon.namespace.modelA.outputs",
                            "seldon.namespace.modelB.outputs",
                        ),
                        tensorsByTopic = mapOf(
                            "seldon.namespace.modelA.outputs" to setOf("tensor1"),
                            "seldon.namespace.modelB.outputs" to setOf("tensor2"),
                        ),
                    ),
                    listOf(
                        "seldon.namespace.modelA.outputs.tensor1",
                        "seldon.namespace.modelB.outputs.tensor2",
                    ),
                ),
                arguments(
                    "tensors override plain topic",
                    makeChainerFor(
                        inputTopic = "seldon.namespace.modelA.outputs",
                        tensors = setOf("tensorA"),
                    ),
                    listOf(
                        "seldon.namespace.modelA.outputs.tensorA",
                        "seldon.namespace.modelA.outputs",
                    ),
                ),
            )

        private fun makeChainerFor(inputTopic: TopicName, tensors: Set<TensorName>?): Chainer =
            Chainer(
                inputTopic = inputTopic,
                tensors = tensors,
                outputTopic = defaultSink,
                pipelineName = defaultPipelineName,
                properties = KafkaProperties(),
                tensorRenaming = emptyMap(),
                kafkaDomainParams = kafkaDomainParams,
                inputTriggerTopics = emptySet(),
                triggerJoinType = ChainerOuterClass.PipelineStepUpdate.PipelineJoinType.Inner,
                triggerTensorsByTopic = emptyMap(),
            )

        private fun makeJoinerFor(
            inputTopics: Set<TopicName>,
            tensorsByTopic: Map<TopicName, Set<TensorName>>?,
        ): Joiner =
            Joiner(
                inputTopics = inputTopics,
                tensorsByTopic = tensorsByTopic,
                outputTopic = defaultSink,
                pipelineName = defaultPipelineName,
                properties = KafkaProperties(),
                tensorRenaming = emptyMap(),
                kafkaDomainParams = kafkaDomainParams,
                joinType = ChainerOuterClass.PipelineStepUpdate.PipelineJoinType.Inner,
                inputTriggerTopics = emptySet(),
                triggerJoinType = ChainerOuterClass.PipelineStepUpdate.PipelineJoinType.Inner,
                triggerTensorsByTopic = emptyMap(),
            )
    }
}

fun Assertion.Builder<Transformer>.isSameTypeAs(other: Transformer) =
    assert("Same type") {
        when {
            it::class == other::class -> pass()
            else -> fail(actual = other::class.simpleName)
        }
    }

fun Assertion.Builder<Transformer>.matches(expected: Transformer) =
    assert("Type and values are the same") {
        when {
            it is Chainer && expected is Chainer -> expect {
                that(it) {
                    get { pipelineName }.isEqualTo(expected.pipelineName)
                    get { inputTopic }.isEqualTo(expected.inputTopic)
                    get { outputTopic }.isEqualTo(expected.outputTopic)
                    get { tensors }.isEqualTo(expected.tensors)
                }
            }
            it is Joiner && expected is Joiner -> expect {
                that(it) {
                    get { pipelineName }.isEqualTo(expected.pipelineName)
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