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
    internal val outerJoin: Boolean,
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
        val nextStream = if (outputTopic.endsWith("outputs")) {
            builder
                .stream(topic, consumerSerde)
                .filterForPipeline(pipelineName)
        } else {
            builder
                .stream(topic, consumerSerde)
                .filterForPipeline(pipelineName)
                .unmarshallInferenceV2()
                .convertToRequests(topic, tensorsByTopic?.get(topic), tensorRenaming)
                .marshallInferenceV2()
        }

        if (outerJoin) {
            val nextPending = pending
                ?.outerJoin(
                    nextStream,
                    ::joinRequests,
                    //JoinWindows.ofTimeDifferenceAndGrace(Duration.ofMillis(1), Duration.ofMillis(1)),
                    // Required because this "fix" causes outer joins to wait for next record to come in if all streams
                    // don't produce a record during grace period. https://issues.apache.org/jira/browse/KAFKA-10847
                    // Also see https://confluentcommunity.slack.com/archives/C6UJNMY67/p1649520904545229?thread_ts=1649324912.542999&cid=C6UJNMY67
                    // Issue created at https://issues.apache.org/jira/browse/KAFKA-13813
                    JoinWindows.of(Duration.ofMillis(1)),
                    joinSerde,
                ) ?: nextStream

            return buildTopology(builder, inputTopics.minus(topic), nextPending)
        } else {
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
    }

    private fun joinRequests(left: ByteArray?, right: ByteArray?): ByteArray {
        if (left == null) {
            return right!!
        }
        if (right == null) {
            return left
        }
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
        logger.info("starting for ($inputTopics) -> ($outputTopic) outerJoin:${outerJoin}")
        streams.start()
    }

    override fun stop() {
        logger.info("stopping joiner ${outputTopic}")
        streams.close()
        // Does not clean up everything see https://issues.apache.org/jira/browse/KAFKA-13787
        streams.cleanUp()
    }

    companion object {
        private val logger = noCoLogger(Joiner::class)
    }
}