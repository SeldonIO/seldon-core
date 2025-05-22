package io.seldon.dataflow.kafka

import io.klogging.noCoLogger
import org.apache.kafka.streams.kstream.Branched
import org.apache.kafka.streams.kstream.KStream
import org.apache.kafka.streams.kstream.Named
import org.apache.kafka.streams.processor.api.FixedKeyProcessor
import org.apache.kafka.streams.processor.api.FixedKeyProcessorContext
import org.apache.kafka.streams.processor.api.FixedKeyRecord
import org.apache.kafka.streams.state.KeyValueStore
import java.util.UUID

const val ERROR_HEADER_KEY = "seldon-pipeline-errors"
const val ERROR_PREFIX = "ERROR:"

const val VISITING_COUNTER_STORE = "visiting-counter-store"
const val VISITING_COUNTER_RETENTION = "120000"

const val VISITING_ERROR_BRANCH = "error"
const val VISITING_DEFAULT_BRANCH = "default"

class VisitingCounterProcessor(
    private val outputTopic: TopicForPipeline,
    private val maxCycles: Int = 2,
) : FixedKeyProcessor<RequestId, TRecord, TRecord> {
    private lateinit var visitingCounterStore: KeyValueStore<String, Int>
    private lateinit var context: FixedKeyProcessorContext<RequestId, TRecord>

    override fun init(context: FixedKeyProcessorContext<RequestId, TRecord>) {
        this.context = context
        this.visitingCounterStore = context.getStateStore(VISITING_COUNTER_STORE) as KeyValueStore<String, Int>
    }

    override fun process(record: FixedKeyRecord<RequestId, TRecord>) {
        val requestId = record.key().toString()
        val compositeKey = "${outputTopic.pipelineName}:${outputTopic.topicName}:$requestId"
        val newCount = (visitingCounterStore.get(compositeKey) ?: 0) + 1
        visitingCounterStore.put(compositeKey, newCount)

        if (newCount > maxCycles) {
            val message = "$ERROR_PREFIX Max cycles ($maxCycles) exceeded for request $requestId in topic $outputTopic"
            logger.warn { message }
            context.forward(
                record
                    .withHeaders(
                        record.headers()
                            .add(
                                ERROR_HEADER_KEY,
                                outputTopic.pipelineName.toByteArray(),
                            ),
                    )
                    .withValue(
                        message.toByteArray(),
                    ),
            )
        } else {
            context.forward(record)
        }
    }

    override fun close() {}

    companion object {
        private val logger = noCoLogger(VisitingCounterProcessor::class)
    }
}

fun isError(value: ByteArray): Boolean {
    return String(value).startsWith(ERROR_PREFIX)
}

fun createVisitingCounterBranches(stream: KStream<RequestId, TRecord>): Pair<KStream<RequestId, TRecord>, KStream<RequestId, TRecord>> {
    val uuid = UUID.randomUUID().toString()
    val branches =
        stream
            .split(Named.`as`("$uuid-"))
            .branch({ _, value -> isError(value) }, Branched.`as`(VISITING_ERROR_BRANCH))
            .defaultBranch(Branched.`as`(VISITING_DEFAULT_BRANCH))

    val errorBranch = requireNotNull(branches["$uuid-$VISITING_ERROR_BRANCH"])
    val defaultBranch = requireNotNull(branches["$uuid-$VISITING_DEFAULT_BRANCH"])
    return Pair(defaultBranch, errorBranch)
}
