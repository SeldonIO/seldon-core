package io.seldon.dataflow.kafka

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
                emptyMap(),
                defaultSink,
                baseKafkaProperties,
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
            KafkaParams(bootstrapServers = "", numCores = 0),
        )

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
                        inputTopics = setOf("seldon.namespace.model1.outputs","seldon.namespace.model2.outputs"),
                        tensors = emptyMap(),
                    ),
                    listOf(
                        "seldon.namespace.modelA.outputs",
                        "seldon.namespace.modelB.outputs",
                    ),
                ),
                arguments(
                    "multiple sources, multiple tensors",
                    makeJoinerFor(
                        inputTopics = setOf("seldon.namespace.model1.outputs","seldon.namespace.model2.outputs"),
                        tensors = mapOf("seldon.namespace.model1.outputs" to setOf("OUTPUT0","OUTPUT1")),
                    ),
                    listOf(
                        "seldon.namespace.modelA.outputs.tensorA",
                        "seldon.namespace.modelB.outputs.tensorB",
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
                tensorMap = emptyMap(),
            )

        private fun makeJoinerFor(inputTopics: Set<TopicName>, tensors: Map<TensorName, Set<TensorName>>): Joiner =
            Joiner(
                inputTopics = inputTopics,
                tensors = tensors,
                outputTopic = defaultSink,
                pipelineName = defaultPipelineName,
                properties = KafkaProperties(),
                tensorMap = emptyMap()
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
            it is Joiner && expected is Joiner -> pass()
            else -> fail(actual = expected)
        }
    }