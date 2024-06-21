/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow.kafka

import io.seldon.mlops.chainer.ChainerOuterClass
import org.apache.kafka.streams.StreamsBuilder
import org.apache.kafka.streams.kstream.JoinWindows
import org.apache.kafka.streams.kstream.KStream
import java.time.Duration

fun addTriggerTopology(
    kafkaDomainParams: KafkaDomainParams,
    builder: StreamsBuilder,
    inputTopics: Set<TopicForPipeline>,
    tensorsByTopic: Map<TopicForPipeline, Set<TensorName>>?,
    joinType: ChainerOuterClass.PipelineStepUpdate.PipelineJoinType,
    lastStream: KStream<RequestId, TRecord>,
    pending: KStream<RequestId, TRecord>? = null,
): KStream<RequestId, TRecord> {
    if (inputTopics.isEmpty()) {
        when (pending) {
            null -> return lastStream
            else ->
                return lastStream
                    .join(
                        pending,
                        ::joinTriggerRequests,
                        JoinWindows.ofTimeDifferenceWithNoGrace(
                            Duration.ofMillis(kafkaDomainParams.joinWindowMillis),
                        ),
                        joinSerde,
                    )
        }
    }

    val topic = inputTopics.first()
    val nextStream =
        builder // TODO possible bug - not all streams will be v2 requests? Maybe v2 responses?
            .stream(topic.topicName, consumerSerde)
            .filterForPipeline(topic.pipelineName)
            .unmarshallInferenceV2Response()
            .convertToRequest(topic.pipelineName, topic.topicName, tensorsByTopic?.get(topic), emptyList())
            // handle cases where there are no tensors we want
            .filter { _, value -> value.inputsList.size != 0 }
            .marshallInferenceV2Request()

    when (joinType) {
        ChainerOuterClass.PipelineStepUpdate.PipelineJoinType.Any -> {
            val nextPending =
                pending
                    ?.outerJoin(
                        nextStream,
                        ::joinTriggerRequests,
                        // JoinWindows.ofTimeDifferenceAndGrace(Duration.ofMillis(1), Duration.ofMillis(1)),
                        // Required because this "fix" causes outer joins to wait for next record to come in if all streams
                        // don't produce a record during grace period. https://issues.apache.org/jira/browse/KAFKA-10847
                        // Also see https://confluentcommunity.slack.com/archives/C6UJNMY67/p1649520904545229?thread_ts=1649324912.542999&cid=C6UJNMY67
                        // Issue created at https://issues.apache.org/jira/browse/KAFKA-13813
                        JoinWindows.of(Duration.ofMillis(1)),
                        joinSerde,
                    ) ?: nextStream

            return addTriggerTopology(
                kafkaDomainParams,
                builder,
                inputTopics.minus(topic),
                tensorsByTopic,
                joinType,
                lastStream,
                nextPending,
            )
        }

        ChainerOuterClass.PipelineStepUpdate.PipelineJoinType.Outer -> {
            val nextPending =
                pending
                    ?.outerJoin(
                        nextStream,
                        ::joinTriggerRequests,
                        // See above for Any case as this will wait until next record comes in before emitting a result after window
                        JoinWindows.ofTimeDifferenceWithNoGrace(
                            Duration.ofMillis(kafkaDomainParams.joinWindowMillis),
                        ),
                        joinSerde,
                    ) ?: nextStream

            return addTriggerTopology(
                kafkaDomainParams,
                builder,
                inputTopics.minus(topic),
                tensorsByTopic,
                joinType,
                lastStream,
                nextPending,
            )
        }

        else -> {
            val nextPending =
                pending
                    ?.join(
                        nextStream,
                        ::joinTriggerRequests,
                        JoinWindows.ofTimeDifferenceWithNoGrace(
                            Duration.ofMillis(kafkaDomainParams.joinWindowMillis),
                        ),
                        joinSerde,
                    ) ?: nextStream

            return addTriggerTopology(
                kafkaDomainParams,
                builder,
                inputTopics.minus(topic),
                tensorsByTopic,
                joinType,
                lastStream,
                nextPending,
            )
        }
    }
}

// For triggers eventually we always want the left item which is the real data returned as its
// <real-data> join <trigger1> join <trigger2> ...
// However for triggers joined to triggers its ok to return anyone that is not null
private fun joinTriggerRequests(
    left: ByteArray?,
    right: ByteArray?,
): ByteArray {
    return left ?: right!!
}
