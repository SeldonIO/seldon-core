package io.seldon.dataflow.kafka

import io.klogging.noCoLogger
import io.seldon.mlops.chainer.ChainerOuterClass
import org.apache.kafka.streams.StreamsBuilder

/**
 * A *step* for a single input stream to a single output stream.
 */
class Chainer(
    builder: StreamsBuilder,
    internal val inputTopic: TopicName,
    internal val outputTopic: TopicName,
    internal val tensors: Set<TensorName>?,
    internal val pipelineName: String,
    internal val tensorRenaming: Map<TensorName, TensorName>,
    internal val batchProperties: ChainerOuterClass.Batch,
    private val kafkaDomainParams: KafkaDomainParams,
    internal val inputTriggerTopics: Set<TopicName>,
    internal val triggerJoinType: ChainerOuterClass.PipelineStepUpdate.PipelineJoinType,
    internal val triggerTensorsByTopic: Map<TopicName, Set<TensorName>>?,
) : PipelineStep {
    init {
        builder.apply {
            if (batchProperties.size > 0) {
                this.addStateStore(BatchProcessor.stateStoreBuilder)
            }
        }

        when (ChainType.create(inputTopic, outputTopic)) {
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
            .stream(inputTopic, consumerSerde)
            .filterForPipeline(pipelineName)
        addTriggerTopology(
            pipelineName,
            kafkaDomainParams,
            builder,
            inputTriggerTopics,
            triggerTensorsByTopic,
            triggerJoinType,
            s1,
            null,
        )
            .headerRemover()
            .to(outputTopic, producerSerde)
    }

    private fun buildInputOutputStream(builder: StreamsBuilder) {
        val s1 = builder
            .stream(inputTopic, consumerSerde)
            .filterForPipeline(pipelineName)
            .unmarshallInferenceV2Request()
            .convertToResponse(inputTopic, tensors, tensorRenaming)
            // handle cases where there are no tensors we want
            .filter { _, value -> value.outputsList.size != 0 }
            .marshallInferenceV2Response()
        addTriggerTopology(
            pipelineName,
            kafkaDomainParams,
            builder,
            inputTriggerTopics,
            triggerTensorsByTopic,
            triggerJoinType,
            s1,
            null,
        )
            .headerRemover()
            .to(outputTopic, producerSerde)
    }

    private fun buildOutputOutputStream(builder: StreamsBuilder) {
        val s1 = builder
            .stream(inputTopic, consumerSerde)
            .filterForPipeline(pipelineName)
            .unmarshallInferenceV2Response()
            .filterResponses(inputTopic, tensors, tensorRenaming)
            // handle cases where there are no tensors we want
            .filter { _, value -> value.outputsList.size != 0 }
            .marshallInferenceV2Response()
        addTriggerTopology(
            pipelineName,
            kafkaDomainParams,
            builder,
            inputTriggerTopics,
            triggerTensorsByTopic,
            triggerJoinType,
            s1,
            null,
        )
            .headerRemover()
            .to(outputTopic, producerSerde)
    }

    private fun buildOutputInputStream(builder: StreamsBuilder) {
        val s1 = builder
            .stream(inputTopic, consumerSerde)
            .filterForPipeline(pipelineName)
            .unmarshallInferenceV2Response()
            .convertToRequest(inputTopic, tensors, tensorRenaming)
            // handle cases where there are no tensors we want
            .filter { _, value -> value.inputsList.size != 0 }
            .batchMessages(batchProperties)
            .marshallInferenceV2Request()
        addTriggerTopology(
            pipelineName,
            kafkaDomainParams,
            builder,
            inputTriggerTopics,
            triggerTensorsByTopic,
            triggerJoinType,
            s1,
            null,
        )
            .headerRemover()
            .to(outputTopic, producerSerde)
    }

    private fun buildInputInputStream(builder: StreamsBuilder) {
        val s1 = builder
            .stream(inputTopic, consumerSerde)
            .filterForPipeline(pipelineName)
            .unmarshallInferenceV2Request()
            .filterRequests(inputTopic, tensors, tensorRenaming)
            // handle cases where there are no tensors we want
            .filter { _, value -> value.inputsList.size != 0 }
            .batchMessages(batchProperties)
            .marshallInferenceV2Request()
        addTriggerTopology(
            pipelineName,
            kafkaDomainParams,
            builder,
            inputTriggerTopics,
            triggerTensorsByTopic,
            triggerJoinType,
            s1,
            null,
        )
            .headerRemover()
            .to(outputTopic, producerSerde)
    }

    companion object {
        private val logger = noCoLogger(Chainer::class)
    }
}