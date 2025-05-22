package io.seldon.dataflow.kafka

import io.klogging.noCoLogger
import org.apache.kafka.streams.processor.api.FixedKeyProcessor
import org.apache.kafka.streams.processor.api.FixedKeyProcessorContext
import org.apache.kafka.streams.processor.api.FixedKeyRecord
import org.apache.kafka.streams.state.KeyValueStore

const val ERROR_HEADER_KEY = "seldon-pipeline-errors"
const val VISITING_COUNTER_STORE = "visiting-counter-store"

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
        val compositeKey = "$outputTopic:$requestId"

        val newCount = (visitingCounterStore.get(compositeKey) ?: 0) + 1
        visitingCounterStore.put(compositeKey, newCount)

        if (newCount > maxCycles) {
            val message = "ERROR: Max cycles ($maxCycles) exceeded for request $requestId in topic $outputTopic"
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
