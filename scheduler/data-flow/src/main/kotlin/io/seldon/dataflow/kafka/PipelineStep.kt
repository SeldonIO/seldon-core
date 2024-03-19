/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow.kafka

import io.seldon.mlops.chainer.ChainerOuterClass.Batch
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineStepUpdate.PipelineJoinType
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineTensorMapping
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineTopic
import org.apache.kafka.streams.StreamsBuilder

interface PipelineStep

data class TopicTensors(
    val topicForPipeline: TopicForPipeline,
    val tensors: Set<TensorName>,
)

data class TopicForPipeline(
    val topicName: TopicName,
    val pipelineName: String,
)

fun stepFor(
    builder: StreamsBuilder,
    pipelineName: String,
    pipelineVersion: Int,
    sources: List<PipelineTopic>,
    triggerSources: List<PipelineTopic>,
    tensorMap: List<PipelineTensorMapping>,
    sink: PipelineTopic,
    joinType: PipelineJoinType,
    triggerJoinType: PipelineJoinType,
    batchProperties: Batch,
    kafkaDomainParams: KafkaDomainParams,
): PipelineStep? {
    val triggerTopicsToTensors = parseTriggers(triggerSources)
    return when (val result = parseSources(sources)) {
        is SourceProjection.Empty -> null
        is SourceProjection.Single -> Chainer(
            builder,
            result.topicForPipeline,
            TopicForPipeline(topicName = sink.topicName, pipelineName = sink.pipelineName),
            null,
            pipelineName,
            pipelineVersion,
            tensorMap,
            batchProperties,
            kafkaDomainParams,
            triggerTopicsToTensors.keys,
            triggerJoinType,
            triggerTopicsToTensors
        )
        is SourceProjection.SingleSubset -> Chainer(
            builder,
            result.topicForPipeline,
            TopicForPipeline(topicName = sink.topicName, pipelineName = sink.pipelineName),
            result.tensors,
            pipelineName,
            pipelineVersion,
            tensorMap,
            batchProperties,
            kafkaDomainParams,
            triggerTopicsToTensors.keys,
            triggerJoinType,
            triggerTopicsToTensors
        )
        is SourceProjection.Many -> Joiner(
            builder,
            result.topicNames,
            TopicForPipeline(topicName = sink.topicName, pipelineName = sink.pipelineName),
            null,
            pipelineName,
            pipelineVersion,
            tensorMap,
            kafkaDomainParams,
            joinType,
            triggerTopicsToTensors.keys,
            triggerJoinType,
            triggerTopicsToTensors
        )
        is SourceProjection.ManySubsets -> Joiner(
            builder,
            result.tensorsByTopic.keys,
            TopicForPipeline(topicName = sink.topicName, pipelineName = sink.pipelineName),
            result.tensorsByTopic,
            pipelineName,
            pipelineVersion,
            tensorMap,
            kafkaDomainParams,
            joinType,
            triggerTopicsToTensors.keys,
            triggerJoinType,
            triggerTopicsToTensors
        )
    }
}


sealed class SourceProjection {
    object Empty : SourceProjection()
    data class Single(val topicForPipeline: TopicForPipeline) : SourceProjection()
    data class SingleSubset(val topicForPipeline: TopicForPipeline, val tensors: Set<TensorName>) : SourceProjection()
    data class Many(val topicNames: Set<TopicForPipeline>) : SourceProjection()
    data class ManySubsets(val tensorsByTopic: Map<TopicForPipeline, Set<TensorName>>) : SourceProjection()
}

fun parseTriggers(sources: List<PipelineTopic>): Map<TopicForPipeline,Set<TensorName>> {
    return sources
        .map { parseSource(it) }
        .groupBy(keySelector = { it.first+":"+it.third }, valueTransform = { it.second })
        .mapValues { it.value.filterNotNull().toSet() }
        .map { TopicTensors(TopicForPipeline(topicName = it.key.split(":")[0], pipelineName = it.key.split(":")[1]), it.value) }
        .associate {it.topicForPipeline to it.tensors }
}

fun parseSources(sources: List<PipelineTopic>): SourceProjection {
    val topicsAndTensors = sources
        .map { parseSource(it) }
        .groupBy(keySelector = { it.first+":"+it.third }, valueTransform = { it.second })
        .mapValues { it.value.filterNotNull().toSet() }
        .map { TopicTensors(TopicForPipeline(topicName = it.key.split(":")[0], pipelineName = it.key.split(":")[1]), it.value) }

    return when {
        topicsAndTensors.isEmpty() -> SourceProjection.Empty
        topicsAndTensors.size == 1 && topicsAndTensors.first().tensors.isEmpty() ->
            SourceProjection.Single(topicsAndTensors.first().topicForPipeline)
        topicsAndTensors.size == 1 ->
            SourceProjection.SingleSubset(
                topicsAndTensors.first().topicForPipeline,
                topicsAndTensors.first().tensors,
            )
        topicsAndTensors.all { it.tensors.isEmpty() } ->
            SourceProjection.Many(topicsAndTensors.map { it.topicForPipeline }.toSet())
        else ->
            SourceProjection.ManySubsets(
                topicsAndTensors.associate { it.topicForPipeline to it.tensors },
            )
    }
}

fun parseSource(source: PipelineTopic): Triple<TopicName, TensorName?, String> {
    return Triple(
        source.topicName,
        if (source.tensor == "") null else source.tensor,
        source.pipelineName,
    )
}