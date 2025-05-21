/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow.kafka

import io.klogging.noCoLogger
import io.seldon.mlops.chainer.ChainerOuterClass
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineTensorMapping
import org.apache.kafka.streams.StreamsBuilder

fun isError(value: ByteArray): Boolean {
    return String(value).contains("ERROR:")
}

/**
 * A *step* for a single input stream to a single output stream.
 */
class Chainer(
    builder: StreamsBuilder,
    internal val inputTopic: TopicForPipeline,
    internal val outputTopic: TopicForPipeline,
    internal val errorTopic: String,
    internal val tensors: Set<TensorName>?,
    internal val pipelineName: String,
    internal val pipelineVersion: String,
    internal val tensorRenaming: List<PipelineTensorMapping>,
    internal val batchProperties: ChainerOuterClass.Batch,
    private val kafkaDomainParams: KafkaDomainParams,
    internal val inputTriggerTopics: Set<TopicForPipeline>,
    internal val triggerJoinType: ChainerOuterClass.PipelineStepUpdate.PipelineJoinType,
    internal val triggerTensorsByTopic: Map<TopicForPipeline, Set<TensorName>>?,
) : PipelineStep {
    init {
        builder.apply {
            if (batchProperties.size > 0) {
                this.addStateStore(BatchProcessor.stateStoreBuilder)
            }
        }

        when (ChainType.create(inputTopic.topicName, outputTopic.topicName)) {
            ChainType.OUTPUT_INPUT -> buildOutputInputStream(builder)
            ChainType.INPUT_INPUT -> buildInputInputStream(builder)
            ChainType.OUTPUT_OUTPUT -> buildOutputOutputStream(builder)
            ChainType.INPUT_OUTPUT -> buildInputOutputStream(builder)
            else -> buildPassThroughStream(builder)
        }

        // TODO - when does K-Streams send an ack?  On consuming or only once a new value has been produced?
        // TODO - wait until streams exists, if it does not already
    }

    private fun buildPassThroughStream(builder: StreamsBuilder) {
        var s1 = builder.stream(inputTopic.topicName, consumerSerde)

        if (inputTopic.topicName != errorTopic) {
            s1 =
                s1.processValues(
                    { CycleTrackingProcessor(errorTopic) },
                    "cycle-store",
                )
            val branches =
                s1.branch(
                    { _, value -> !isError(value) },
                    { _, value -> isError(value) },
                )

            branches[1]
                .to(errorTopic, producerSerde)

            addTriggerTopology(
                kafkaDomainParams,
                builder,
                inputTriggerTopics,
                triggerTensorsByTopic,
                triggerJoinType,
                branches[0],
                null,
            )
                .headerRemover()
                .headerSetter(pipelineName, pipelineVersion)
                .to(outputTopic.topicName, producerSerde)
        } else {
            addTriggerTopology(
                kafkaDomainParams,
                builder,
                inputTriggerTopics,
                triggerTensorsByTopic,
                triggerJoinType,
                s1,
                null,
            )
                .headerRemover()
                .headerSetter(pipelineName, pipelineVersion)
                .to(outputTopic.topicName, producerSerde)
        }
    }

    private fun buildInputOutputStream(builder: StreamsBuilder) {
        val s1 =
            builder
                .stream(inputTopic.topicName, consumerSerde)
                .processValues({ CycleTrackingProcessor(errorTopic) }, "cycle-store")

        val branches =
            s1.branch(
                { _, value -> !isError(value) },
                { _, value -> isError(value) },
            )

        branches[0]
            .filterForPipeline(inputTopic.pipelineName)
            .unmarshallInferenceV2Request()
            .convertToResponse(inputTopic.pipelineName, inputTopic.topicName, tensors, tensorRenaming)
            // handle cases where there are no tensors we want
            .filter { _, value -> value.outputsList.size != 0 }
            .marshallInferenceV2Response()
        branches[1]
            .to(errorTopic, producerSerde)

        addTriggerTopology(
            kafkaDomainParams,
            builder,
            inputTriggerTopics,
            triggerTensorsByTopic,
            triggerJoinType,
            branches[0],
            null,
        )
            .headerRemover()
            .headerSetter(pipelineName, pipelineVersion)
            .to(outputTopic.topicName, producerSerde)
    }

    private fun buildOutputOutputStream(builder: StreamsBuilder) {
        val s1 =
            builder
                .stream(inputTopic.topicName, consumerSerde)
                .processValues({ CycleTrackingProcessor(errorTopic) }, "cycle-store")

        val branches =
            s1.branch(
                { _, value -> !isError(value) },
                { _, value -> isError(value) },
            )

        branches[0]
            .filterForPipeline(inputTopic.pipelineName)
            .unmarshallInferenceV2Response()
            .filterResponses(inputTopic.pipelineName, inputTopic.topicName, tensors, tensorRenaming)
            // handle cases where there are no tensors we want
            .filter { _, value -> value.outputsList.size != 0 }
            .marshallInferenceV2Response()
        branches[1]
            .to(errorTopic, producerSerde)

        addTriggerTopology(
            kafkaDomainParams,
            builder,
            inputTriggerTopics,
            triggerTensorsByTopic,
            triggerJoinType,
            branches[0],
            null,
        )
            .headerRemover()
            .headerSetter(pipelineName, pipelineVersion)
            .to(outputTopic.topicName, producerSerde)
    }

    private fun buildOutputInputStream(builder: StreamsBuilder) {
        val s1 =
            builder
                .stream(inputTopic.topicName, consumerSerde)
                .processValues({ CycleTrackingProcessor(errorTopic) }, "cycle-store")

        val branches =
            s1.branch(
                { _, value -> !isError(value) },
                { _, value -> isError(value) },
            )

        branches[0]
            .filterForPipeline(inputTopic.pipelineName)
            .unmarshallInferenceV2Response()
            .convertToRequest(inputTopic.pipelineName, inputTopic.topicName, tensors, tensorRenaming)
            // handle cases where there are no tensors we want
            .filter { _, value -> value.inputsList.size != 0 }
            .batchMessages(batchProperties)
            .marshallInferenceV2Request()
        branches[1]
            .to(errorTopic, producerSerde)

        addTriggerTopology(
            kafkaDomainParams,
            builder,
            inputTriggerTopics,
            triggerTensorsByTopic,
            triggerJoinType,
            branches[0],
            null,
        )
            .headerRemover()
            .headerSetter(pipelineName, pipelineVersion)
            .to(outputTopic.topicName, producerSerde)
    }

    private fun buildInputInputStream(builder: StreamsBuilder) {
        val s1 =
            builder
                .stream(inputTopic.topicName, consumerSerde)
                .processValues({ CycleTrackingProcessor(errorTopic) }, "cycle-store")

        val branches =
            s1.branch(
                { _, value -> !isError(value) },
                { _, value -> isError(value) },
            )

        branches[0]
            .filterForPipeline(inputTopic.pipelineName)
            .unmarshallInferenceV2Request()
            .filterRequests(inputTopic.pipelineName, inputTopic.topicName, tensors, tensorRenaming)
            // handle cases where there are no tensors we want
            .filter { _, value -> value.inputsList.size != 0 }
            .batchMessages(batchProperties)
            .marshallInferenceV2Request()
        branches[1]
            .to(errorTopic, producerSerde)

        addTriggerTopology(
            kafkaDomainParams,
            builder,
            inputTriggerTopics,
            triggerTensorsByTopic,
            triggerJoinType,
            branches[0],
            null,
        )
            .headerRemover()
            .headerSetter(pipelineName, pipelineVersion)
            .to(outputTopic.topicName, producerSerde)
    }

    companion object {
        private val logger = noCoLogger(Chainer::class)
    }
}
