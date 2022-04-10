package io.seldon.dataflow.kafka

import io.klogging.noCoLogger
import org.apache.kafka.streams.KafkaStreams
import org.apache.kafka.streams.StreamsBuilder

/**
 * A *transformer* for a single input stream to a single output stream.
 */
class Chainer(
    properties: KafkaProperties,
    internal val inputTopic: TopicName,
    internal val outputTopic: TopicName,
    internal val tensors: Set<TensorName>?,
    internal val pipelineName: String,
    internal val tensorRenaming: Map<TensorName, TensorName>,
    private val kafkaDomainParams: KafkaDomainParams,
) : Transformer {
    private val streams: KafkaStreams by lazy {
        val builder = StreamsBuilder()

        if (inputTopic.endsWith("outputs") && outputTopic.endsWith("inputs")) {
            builder
                .stream(inputTopic, consumerSerde)
                .filterForPipeline(pipelineName)
                .unmarshallInferenceV2()
                .convertToRequests(inputTopic, tensors, tensorRenaming)
                // handle cases where there are no tensors we want
                .filter { _, value -> value.inputsList.size != 0}
                .marshallInferenceV2()
                .to(outputTopic, producerSerde)
        } else {
            builder
                .stream(inputTopic, consumerSerde)
                .filterForPipeline(pipelineName)
                .to(outputTopic, producerSerde)
        }

        // TODO - when does K-Streams send an ack?  On consuming or only once a new value has been produced?
        // TODO - wait until streams exists, if it does not already

        KafkaStreams(builder.build(), properties)
    }

    override fun start() {
        if (kafkaDomainParams.useCleanState) {
            streams.cleanUp()
        }
        logger.info("starting for ($inputTopic) -> ($outputTopic)")
        streams.start()
    }

    override fun stop() {
        logger.info("stopping chainer ${inputTopic}->${outputTopic}")
        streams.close()
        streams.cleanUp()
    }

    companion object {
        private val logger = noCoLogger(Chainer::class)
    }
}