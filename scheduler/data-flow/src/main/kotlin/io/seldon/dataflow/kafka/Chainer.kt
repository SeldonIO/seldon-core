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
import io.seldon.mlops.chainer.ChainerOuterClass
import org.apache.kafka.streams.StreamsBuilder

/**
 * A *step* for a single input stream to a single output stream.
 */
class Chainer(
    builder: StreamsBuilder,
    internal val inputTopic: TopicForPipeline,
    internal val outputTopic: TopicForPipeline,
    internal val tensors: Set<TensorName>?,
    internal val pipelineName: String,
    internal val tensorRenaming: Map<TensorName, TensorName>,
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
        val s1 = builder
            .stream(inputTopic.topicName, consumerSerde)
            .filterForPipeline(inputTopic.pipelineName)
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
            .headerSetter(pipelineName)
            .to(outputTopic.topicName, producerSerde)
    }

    private fun buildInputOutputStream(builder: StreamsBuilder) {
        val s1 = builder
            .stream(inputTopic.topicName, consumerSerde)
            .filterForPipeline(inputTopic.pipelineName)
            .unmarshallInferenceV2Request()
            .convertToResponse(inputTopic.topicName, tensors, tensorRenaming)
            // handle cases where there are no tensors we want
            .filter { _, value -> value.outputsList.size != 0 }
            .marshallInferenceV2Response()
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
            .headerSetter(pipelineName)
            .to(outputTopic.topicName, producerSerde)
    }

    private fun buildOutputOutputStream(builder: StreamsBuilder) {
        val s1 = builder
            .stream(inputTopic.topicName, consumerSerde)
            .filterForPipeline(inputTopic.pipelineName)
            .unmarshallInferenceV2Response()
            .filterResponses(inputTopic.topicName, tensors, tensorRenaming)
            // handle cases where there are no tensors we want
            .filter { _, value -> value.outputsList.size != 0 }
            .marshallInferenceV2Response()
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
            .headerSetter(pipelineName)
            .to(outputTopic.topicName, producerSerde)
    }

    private fun buildOutputInputStream(builder: StreamsBuilder) {
        val s1 = builder
            .stream(inputTopic.topicName, consumerSerde)
            .filterForPipeline(inputTopic.pipelineName)
            .unmarshallInferenceV2Response()
            .convertToRequest(inputTopic.topicName, tensors, tensorRenaming)
            // handle cases where there are no tensors we want
            .filter { _, value -> value.inputsList.size != 0 }
            .batchMessages(batchProperties)
            .marshallInferenceV2Request()
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
            .headerSetter(pipelineName)
            .to(outputTopic.topicName, producerSerde)
    }

    private fun buildInputInputStream(builder: StreamsBuilder) {
        val s1 = builder
            .stream(inputTopic.topicName, consumerSerde)
            .filterForPipeline(inputTopic.pipelineName)
            .unmarshallInferenceV2Request()
            .filterRequests(inputTopic.topicName, tensors, tensorRenaming)
            // handle cases where there are no tensors we want
            .filter { _, value -> value.inputsList.size != 0 }
            .batchMessages(batchProperties)
            .marshallInferenceV2Request()
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
            .headerSetter(pipelineName)
            .to(outputTopic.topicName, producerSerde)
    }

    companion object {
        private val logger = noCoLogger(Chainer::class)
    }
}