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

import io.seldon.mlops.chainer.ChainerOuterClass.PipelineStepUpdate
import org.apache.kafka.clients.admin.Admin
import org.apache.kafka.clients.admin.NewTopic
import org.apache.kafka.common.KafkaFuture
import org.apache.kafka.common.errors.TopicExistsException
import java.util.concurrent.ExecutionException
import io.klogging.logger as coLogger

class KafkaAdmin(
    adminConfig: KafkaAdminProperties,
    private val streamsConfig: KafkaStreamsParams,
) {
    private val adminClient = Admin.create(adminConfig)

    suspend fun ensureTopicsExist(
        steps: List<PipelineStepUpdate>,
    ) {
        steps
            .flatMap { step -> step.sourcesList + step.sink + step.triggersList }
            .map { topicName -> parseSource(topicName).first }
            .toSet()
            .also {
                logger.info("Topics found are $it")
            }
            .map { topicName ->
                NewTopic(
                    topicName,
                    streamsConfig.numPartitions,
                    streamsConfig.replicationFactor.toShort(),
                ).configs(
                    kafkaTopicConfig(
                        streamsConfig.maxMessageSizeBytes,
                    ),
                )
            }
            .run { adminClient.createTopics(this) }
            .values()
            .also { topicCreations ->
                topicCreations.entries.forEach { creationResult ->
                    awaitKafkaResult(creationResult)
                }
            }
    }

    private suspend fun awaitKafkaResult(result: Map.Entry<String, KafkaFuture<Void>>) {
        try {
            result.value.get()
            logger.info("Topic created ${result.key}")
        } catch (e: ExecutionException) {
            if (e.cause is TopicExistsException) {
                logger.info("Topic already exists ${result.key}")
            } else {
                throw e
            }
        }
    }

    companion object {
        private val logger = coLogger(KafkaAdmin::class)
    }
}
