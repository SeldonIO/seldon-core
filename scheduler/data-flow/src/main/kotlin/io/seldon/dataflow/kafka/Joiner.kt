package io.seldon.dataflow.kafka

import io.klogging.noCoLogger
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineStepUpdate.PipelineJoinType
import io.seldon.mlops.inference.v2.V2Dataplane
import org.apache.kafka.streams.StreamsBuilder
import org.apache.kafka.streams.kstream.JoinWindows
import org.apache.kafka.streams.kstream.KStream
import java.time.Duration

/**
 * A *step* which joins multiple input streams into a single output stream.
 */
class Joiner(
    builder: StreamsBuilder,
    internal val inputTopics: Set<TopicName>,
    internal val outputTopic: TopicName,
    internal val tensorsByTopic: Map<TopicName, Set<TensorName>>?,
    internal val pipelineName: String,
    internal val tensorRenaming: Map<TensorName, TensorName>,
    internal val kafkaDomainParams: KafkaDomainParams,
    internal val joinType: PipelineJoinType,
    internal val inputTriggerTopics: Set<TopicName>,
    internal val triggerJoinType: PipelineJoinType,
    internal val triggerTensorsByTopic: Map<TopicName, Set<TensorName>>?,
) : PipelineStep {
    init {
        val dataStream = buildTopology(builder, inputTopics)
        addTriggerTopology(
            pipelineName,
            kafkaDomainParams,
            builder,
            inputTriggerTopics,
            triggerTensorsByTopic,
            triggerJoinType,
            dataStream,
            null
        )
            .headerRemover()
            .to(outputTopic, producerSerde)
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

        val chainType = ChainType.create(topic, outputTopic)
        logger.info("Creating stream ${chainType} for ${topic}->${outputTopic}")
        val nextStream = when (chainType) {
            ChainType.OUTPUT_INPUT -> buildOutputInputStream(topic, builder)
            ChainType.INPUT_INPUT -> buildInputInputStream(topic, builder)
            ChainType.OUTPUT_OUTPUT -> buildOutputOutputStream(topic, builder)
            ChainType.INPUT_OUTPUT -> buildInputOutputStream(topic, builder)
            else -> buildPassThroughStream(topic, builder)
        }
        val payloadJoiner = when (chainType) {
            ChainType.OUTPUT_INPUT, ChainType.INPUT_INPUT -> ::joinRequests
            ChainType.OUTPUT_OUTPUT, ChainType.INPUT_OUTPUT -> ::joinResponses
            else -> throw Exception("Can't join custom data")
        }

        when (joinType) {
            PipelineJoinType.Any -> {
                val nextPending = pending
                    ?.outerJoin(
                        nextStream,
                        payloadJoiner,
                        //JoinWindows.ofTimeDifferenceAndGrace(Duration.ofMillis(1), Duration.ofMillis(1)),
                        // Required because this "fix" causes outer joins to wait for next record to come in if all streams
                        // don't produce a record during grace period. https://issues.apache.org/jira/browse/KAFKA-10847
                        // Also see https://confluentcommunity.slack.com/archives/C6UJNMY67/p1649520904545229?thread_ts=1649324912.542999&cid=C6UJNMY67
                        // Issue created at https://issues.apache.org/jira/browse/KAFKA-13813
                        JoinWindows.of(Duration.ofMillis(1)),
                        joinSerde,
                    ) ?: nextStream


                return buildTopology(builder, inputTopics.minus(topic), nextPending)
            }

            PipelineJoinType.Outer -> {
                val nextPending = pending
                    ?.outerJoin(
                        nextStream,
                        payloadJoiner,
                        // See above for Any case as this will wait until next record comes in before emitting a result after window
                        JoinWindows.ofTimeDifferenceWithNoGrace(
                            Duration.ofMillis(kafkaDomainParams.joinWindowMillis),
                        ),
                        joinSerde,
                    ) ?: nextStream


                return buildTopology(builder, inputTopics.minus(topic), nextPending)
            }

            else -> {
                val nextPending = pending
                    ?.join(
                        nextStream,
                        payloadJoiner,
                        JoinWindows.ofTimeDifferenceWithNoGrace(
                            Duration.ofMillis(kafkaDomainParams.joinWindowMillis),
                        ),
                        joinSerde,
                    ) ?: nextStream

                return buildTopology(builder, inputTopics.minus(topic), nextPending)
            }
        }
    }

    private fun buildPassThroughStream(topic: String, builder: StreamsBuilder): KStream<RequestId, TRecord> {
        return builder
            .stream(topic, consumerSerde)
            .filterForPipeline(pipelineName)
    }

    private fun buildInputOutputStream(topic: String, builder: StreamsBuilder): KStream<RequestId, TRecord> {
        return builder
            .stream(topic, consumerSerde)
            .filterForPipeline(pipelineName)
            .unmarshallInferenceV2Request()
            .convertToResponse(topic, tensorsByTopic?.get(topic), tensorRenaming)
            // handle cases where there are no tensors we want
            .filter { _, value -> value.outputsList.size != 0 }
            .marshallInferenceV2Response()
    }

    private fun buildOutputOutputStream(topic: String, builder: StreamsBuilder): KStream<RequestId, TRecord> {
        return builder
            .stream(topic, consumerSerde)
            .filterForPipeline(pipelineName)
            .unmarshallInferenceV2Response()
            .filterResponses(topic, tensorsByTopic?.get(topic), tensorRenaming)
            // handle cases where there are no tensors we want
            .filter { _, value -> value.outputsList.size != 0 }
            .marshallInferenceV2Response()
    }

    private fun buildOutputInputStream(topic: String, builder: StreamsBuilder): KStream<RequestId, TRecord> {
        return builder
            .stream(topic, consumerSerde)
            .filterForPipeline(pipelineName)
            .unmarshallInferenceV2Response()
            .convertToRequest(topic, tensorsByTopic?.get(topic), tensorRenaming)
            .marshallInferenceV2Request()
    }

    private fun buildInputInputStream(topic: String, builder: StreamsBuilder): KStream<RequestId, TRecord> {
        return builder
            .stream(topic, consumerSerde)
            .filterForPipeline(pipelineName)
            .unmarshallInferenceV2Request()
            .filterRequests(topic, tensorsByTopic?.get(topic), tensorRenaming)
            // handle cases where there are no tensors we want
            .filter { _, value -> value.inputsList.size != 0 }
            .marshallInferenceV2Request()
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

    private fun joinResponses(left: ByteArray?, right: ByteArray?): ByteArray {
        if (left == null) {
            return right!!
        }
        if (right == null) {
            return left
        }
        val leftResponse = V2Dataplane.ModelInferResponse.parseFrom(left)
        val rightResponse = V2Dataplane.ModelInferResponse.parseFrom(right)
        val response = V2Dataplane.ModelInferResponse
            .newBuilder()
            .setId(leftResponse.id)
            .putAllParameters(leftResponse.parametersMap)
            .addAllOutputs(leftResponse.outputsList)
            .addAllOutputs(rightResponse.outputsList)
            .addAllRawOutputContents(leftResponse.rawOutputContentsList)
            .addAllRawOutputContents(rightResponse.rawOutputContentsList)
            .build()
        return response.toByteArray()
    }

    companion object {
        private val logger = noCoLogger(Joiner::class)
    }
}