/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow.kafka

import com.github.michaelbull.retry.ContinueRetrying
import com.github.michaelbull.retry.policy.RetryPolicy
import com.github.michaelbull.retry.policy.constantDelay
import com.github.michaelbull.retry.policy.limitAttempts
import com.github.michaelbull.retry.policy.plus
import com.github.michaelbull.retry.retry
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineStepUpdate
import org.apache.kafka.clients.admin.Admin
import org.apache.kafka.clients.admin.CreateTopicsOptions
import org.apache.kafka.clients.admin.NewTopic
import org.apache.kafka.common.KafkaFuture
import org.apache.kafka.common.errors.TimeoutException
import org.apache.kafka.common.errors.TopicExistsException
import org.apache.kafka.common.errors.UnknownTopicOrPartitionException
import java.util.concurrent.ExecutionException
import java.util.concurrent.TimeUnit
import io.klogging.logger as coLogger

class KafkaAdmin(
    adminConfig: KafkaAdminProperties,
    private val streamsConfig: KafkaStreamsParams,
) {
    private val adminClient = Admin.create(adminConfig)

    suspend fun ensureTopicsExist(
        steps: List<PipelineStepUpdate>,
    ) : Exception? {
        val missingTopicRetryPolicy: RetryPolicy<Throwable> = {
            when (reason) {
                is TimeoutException,
                is UnknownTopicOrPartitionException -> ContinueRetrying
                else -> {
                    logger.warn("ignoring exception while waiting for topic creation: ${reason.message}")
                    ContinueRetrying
                }
            }
        }

        try {
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
                .run {
                    adminClient.createTopics(this, CreateTopicsOptions().timeoutMs(60_000))
                }
                .values()
                .also { topicCreations ->
                    logger.info("Waiting for kafka topic creation")
                    // We repeatedly attempt to describe all topics as a way of blocking until they exist at least on
                    // one broker. This is because the call to createTopics above returns before topics can actually
                    // be subscribed to.
                    retry(missingTopicRetryPolicy + limitAttempts(60) + constantDelay(delayMillis = 1000L)) {
                        logger.debug("Still waiting for all topics to be created...")
                        adminClient.describeTopics(topicCreations.keys).allTopicNames().get(500, TimeUnit.MILLISECONDS)
                    }
                }
        } catch (e: Exception) {
            // we catch all exceptions here and return them instead, because we want to handle
            // errors as part of programming logic, instead of them bubbling up to the scheduler
            // subscription event loop. This way, errors for one pipeline don't interfere in the
            // execution of others.
            return e
        }

        logger.info("All topics created")
        return null
    }

    companion object {
        private val logger = coLogger(KafkaAdmin::class)
    }
}
