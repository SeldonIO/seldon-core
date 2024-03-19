/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow.kafka

import io.klogging.noCoLogger
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineStepUpdate.PipelineJoinType
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineTensorMapping
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
    internal val inputTopics: Set<TopicForPipeline>,
    internal val outputTopic: TopicForPipeline,
    internal val tensorsByTopic: Map<TopicForPipeline, Set<TensorName>>?,
    internal val pipelineName: String,
    internal val pipelineVersion: Int,
    internal val tensorRenaming: List<PipelineTensorMapping>,
    internal val kafkaDomainParams: KafkaDomainParams,
    internal val joinType: PipelineJoinType,
    internal val inputTriggerTopics: Set<TopicForPipeline>,
    internal val triggerJoinType: PipelineJoinType,
    internal val triggerTensorsByTopic: Map<TopicForPipeline, Set<TensorName>>?,
) : PipelineStep {
    init {
        val dataStream = buildTopology(builder, inputTopics)
        addTriggerTopology(
            kafkaDomainParams,
            builder,
            inputTriggerTopics,
            triggerTensorsByTopic,
            triggerJoinType,
            dataStream,
            null
        )
            .headerRemover()
            .headerSetter(pipelineName, pipelineVersion)
            .to(outputTopic.topicName, producerSerde)
    }

    private fun buildTopology(
        builder: StreamsBuilder,
        inputTopics: Set<TopicForPipeline>,
        pending: KStream<RequestId, TRecord>? = null,
    ): KStream<RequestId, TRecord> {
        if (inputTopics.isEmpty()) {
            when (pending) {
                null -> throw IllegalArgumentException("cannot join zero streams")
                else -> return pending
            }
        }

        val topic = inputTopics.first()

        val chainType = ChainType.create(topic.topicName, outputTopic.topicName)
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

    private fun buildPassThroughStream(topic: TopicForPipeline, builder: StreamsBuilder): KStream<RequestId, TRecord> {
        return builder
            .stream(topic.topicName, consumerSerde)
            .filterForPipeline(topic.pipelineName)
    }

    private fun buildInputOutputStream(topic: TopicForPipeline, builder: StreamsBuilder): KStream<RequestId, TRecord> {
        return builder
            .stream(topic.topicName, consumerSerde)
            .filterForPipeline(topic.pipelineName)
            .unmarshallInferenceV2Request()
            .convertToResponse(topic.pipelineName, topic.topicName, tensorsByTopic?.get(topic), tensorRenaming)
            // handle cases where there are no tensors we want
            .filter { _, value -> value.outputsList.size != 0 }
            .marshallInferenceV2Response()
    }

    private fun buildOutputOutputStream(topic: TopicForPipeline, builder: StreamsBuilder): KStream<RequestId, TRecord> {
        return builder
            .stream(topic.topicName, consumerSerde)
            .filterForPipeline(topic.pipelineName)
            .unmarshallInferenceV2Response()
            .filterResponses(topic.pipelineName, topic.topicName, tensorsByTopic?.get(topic), tensorRenaming)
            // handle cases where there are no tensors we want
            .filter { _, value -> value.outputsList.size != 0 }
            .marshallInferenceV2Response()
    }

    private fun buildOutputInputStream(topic: TopicForPipeline, builder: StreamsBuilder): KStream<RequestId, TRecord> {
        return builder
            .stream(topic.topicName, consumerSerde)
            .filterForPipeline(topic.pipelineName)
            .unmarshallInferenceV2Response()
            .convertToRequest(topic.pipelineName, topic.topicName, tensorsByTopic?.get(topic), tensorRenaming)
            // handle cases where there are no tensors we want
            .filter { _, value -> value.inputsList.size != 0 }
            .marshallInferenceV2Request()
    }

    private fun buildInputInputStream(topic: TopicForPipeline, builder: StreamsBuilder): KStream<RequestId, TRecord> {
        return builder
            .stream(topic.topicName, consumerSerde)
            .filterForPipeline(topic.pipelineName)
            .unmarshallInferenceV2Request()
            .filterRequests(topic.pipelineName,topic.topicName, tensorsByTopic?.get(topic), tensorRenaming)
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
        var leftRequest = V2Dataplane.ModelInferRequest.parseFrom(left)
        var rightRequest = V2Dataplane.ModelInferRequest.parseFrom(right)
        if (leftRequest.rawInputContentsCount > 0 && rightRequest.rawInputContentsCount == 0) {
            rightRequest = rightRequest.withBinaryContents()
        } else if (rightRequest.rawInputContentsCount > 0 && leftRequest.rawInputContentsCount == 0) {
            leftRequest = leftRequest.withBinaryContents()
        }
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
        var leftResponse = V2Dataplane.ModelInferResponse.parseFrom(left)
        var rightResponse = V2Dataplane.ModelInferResponse.parseFrom(right)
        if (leftResponse.rawOutputContentsCount > 0 && rightResponse.rawOutputContentsCount == 0) {
            rightResponse = rightResponse.withBinaryContents()
        } else if (rightResponse.rawOutputContentsCount > 0 && leftResponse.rawOutputContentsCount == 0) {
            leftResponse = leftResponse.withBinaryContents()
        }
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