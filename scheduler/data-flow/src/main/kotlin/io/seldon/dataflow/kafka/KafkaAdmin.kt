package io.seldon.dataflow.kafka

import io.seldon.dataflow.parallel
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineStepUpdate
import kotlinx.coroutines.flow.asFlow
import kotlinx.coroutines.runBlocking
import org.apache.kafka.clients.admin.Admin
import org.apache.kafka.clients.admin.NewTopic
import org.apache.kafka.common.KafkaFuture
import org.apache.kafka.common.errors.TopicExistsException
import java.util.concurrent.ExecutionException
import io.klogging.logger as coLogger

class KafkaAdmin(kafkaProperties: KafkaProperties) {
    private val adminClient = Admin.create(kafkaProperties)

    suspend fun ensureTopicsExist(
        steps: List<PipelineStepUpdate>,
    ) {
        steps
            .flatMap { step -> step.sourcesList + step.sink }
            .map { topicName -> parseSource(topicName).first }
            .toSet()
            .also {
                logger.info("Topics found are $it")
            }
            .map { topicName -> NewTopic(topicName, 1, 1) }
            .run { adminClient.createTopics(this) }
            .values()
            .also { topicCreations ->
                runBlocking {
                    topicCreations.entries
                        .asFlow()
                        .parallel(this) { creationResult ->
                            awaitKafkaResult(creationResult)
                        }
                }
            }
    }

    private suspend fun awaitKafkaResult(result: Map.Entry<String, KafkaFuture<Void>>) {
        try {
            result.value.get()
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