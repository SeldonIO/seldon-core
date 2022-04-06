package io.seldon.dataflow.kafka

import io.klogging.noCoLogger
import io.seldon.mlops.inference.v2.V2Dataplane
import org.apache.kafka.streams.KafkaStreams
import org.apache.kafka.streams.StreamsBuilder
import org.apache.kafka.streams.kstream.*
import java.time.Duration

/**
 * A *transformer* which joins multiple input streams into a single output stream.
 */
class Joiner(
    properties: KafkaProperties,
    internal val inputTopics: Set<TopicName>,
    internal val outputTopic: TopicName,
    internal val tensorsByTopic: Map<TopicName, Set<TensorName>>?,
    internal val pipelineName: String,
    internal val tensorRenaming: Map<TensorName, TensorName>,
    internal val kafkaDomainParams: KafkaDomainParams,
) : Transformer {
    private val streams: KafkaStreams by lazy {
        val builder = StreamsBuilder()
        buildTopology(builder, inputTopics).to(outputTopic, producerSerde)
        KafkaStreams(builder.build(), properties)
    }

    private fun buildTopology(
        builder: StreamsBuilder,
        inputTopics: Set<TopicName>,
        pending: KStream<RequestId, TRecord>? = null,
    ): KStream<RequestId, TRecord> {
        if (inputTopics.isEmpty()) {
            when (pending) {
                null -> throw IllegalArgumentException("cannot join zero streams")
                else -> return pending
            }
        }

        val topic = inputTopics.first()
        val nextStream = builder
            .stream(topic, consumerSerde)
            .filterForPipeline(pipelineName)
            .unmarshallInferenceV2()
            .convertToRequests(topic, tensorsByTopic?.get(topic), tensorRenaming)
            .marshallInferenceV2()

        val nextPending = pending
            ?.join(
                nextStream,
                ::joinRequests,
                JoinWindows.ofTimeDifferenceWithNoGrace(
                    Duration.ofMillis(kafkaDomainParams.joinWindowMillis),
                ),
                joinSerde,
            ) ?: nextStream

        return buildTopology(builder, inputTopics.minus(topic), nextPending)
    }

    private fun joinRequests(left: ByteArray, right: ByteArray): ByteArray {
        val leftRequest = V2Dataplane.ModelInferRequest.parseFrom(left)
        val rightRequest = V2Dataplane.ModelInferRequest.parseFrom(right)
        val request = V2Dataplane.ModelInferRequest
            .newBuilder()
            .setId(leftRequest.id)
            .putAllParameters(leftRequest.parametersMap)
            .addAllInputs(leftRequest.inputsList)
            .addAllInputs(rightRequest.inputsList)
            .addAllRawInputContents(leftRequest.rawInputContentsList)
            .addAllRawInputContents(rightRequest.rawInputContentsList)
            .build()
        return request.toByteArray()
    }

    override fun start() {
        if (kafkaDomainParams.useCleanState) {
            streams.cleanUp()
        }
        logger.info("starting for ($inputTopics) -> ($outputTopic)")
        streams.start()
    }

    override fun stop() {
        logger.info("stopping joiner")
        streams.close()
    }

    companion object {
        private val logger = noCoLogger(Joiner::class)
    }
}