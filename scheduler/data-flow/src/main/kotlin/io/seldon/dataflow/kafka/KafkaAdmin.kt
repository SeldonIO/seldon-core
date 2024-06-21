/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow.kafka

import com.github.michaelbull.retry.policy.constantDelay
import com.github.michaelbull.retry.policy.continueIf
import com.github.michaelbull.retry.policy.plus
import com.github.michaelbull.retry.policy.stopAtAttempts
import com.github.michaelbull.retry.retry
import io.klogging.noCoLogger
import io.seldon.mlops.chainer.ChainerOuterClass.PipelineStepUpdate
import org.apache.kafka.clients.admin.Admin
import org.apache.kafka.clients.admin.CreateTopicsOptions
import org.apache.kafka.clients.admin.NewTopic
import org.apache.kafka.common.errors.TimeoutException
import org.apache.kafka.common.errors.UnknownTopicOrPartitionException
import java.util.concurrent.TimeUnit
import io.klogging.logger as coLogger

class KafkaAdmin(
    adminConfig: KafkaAdminProperties,
    private val streamsConfig: KafkaStreamsParams,
    private val topicWaitRetryParams: TopicWaitRetryParams,
) {
    private val adminClient = Admin.create(adminConfig)

    suspend fun ensureTopicsExist(steps: List<PipelineStepUpdate>): Exception? {
        val missingTopicRetryPolicy =
            continueIf<Throwable> { (failure) ->
                when (failure) {
                    is TimeoutException,
                    is UnknownTopicOrPartitionException,
                    -> true
                    else -> {
                        // We log here for dev purposes, to gather other kinds of exceptions that occur. In time, we should
                        // collate those and decide which are permanent errors. For permanent errors, it would be worth
                        // stopping the retries and returning false.
                        noCoLogger.warn("ignoring exception while waiting for topic creation: ${failure.message}")
                        true
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
                    adminClient.createTopics(
                        this,
                        CreateTopicsOptions().timeoutMs(topicWaitRetryParams.createTimeoutMillis),
                    )
                }
                .values()
                .also { topicCreations ->
                    logger.info("Waiting for kafka topic creation")
                    // We repeatedly attempt to describe all topics as a way of blocking until they exist at least on
                    // one broker. This is because the call to createTopics above returns before topics can actually
                    // be subscribed to.
                    retry(
                        missingTopicRetryPolicy + stopAtAttempts(topicWaitRetryParams.describeRetries) +
                            constantDelay(
                                topicWaitRetryParams.describeRetryDelayMillis,
                            ),
                    ) {
                        logger.debug("Still waiting for all topics to be created...")
                        // the KafkaFuture retrieved via .allTopicNames() only succeeds if all the topic
                        // descriptions succeed, so there is no need to check topic descriptions individually
                        adminClient.describeTopics(topicCreations.keys).allTopicNames()
                            .get(topicWaitRetryParams.describeTimeoutMillis, TimeUnit.MILLISECONDS)
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
        private val noCoLogger = noCoLogger(KafkaAdmin::class)
    }
}
