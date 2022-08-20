package io.seldon.dataflow.kafka

import io.klogging.noCoLogger
import org.apache.kafka.streams.KafkaStreams
import org.apache.kafka.streams.KafkaStreams.StateListener
import java.util.concurrent.CountDownLatch

typealias PipelineId = String

data class PipelineMetadata(
    val id: PipelineId,
    val name: String,
    val version: Int,
)

// TODO - move pipeline creation much more inside this class, away from pipeline subscriber
class PipelineTopology(
    private val metadata: PipelineMetadata,
    private val steps: List<PipelineStep>,
    private val streams: KafkaStreams,
) : StateListener {
    private val latch = CountDownLatch(1)

    fun start(clean: Boolean) {
        if (clean) {
            streams.cleanUp()
        }
        streams.setStateListener(this)
        streams.start()

        // Do not allow pipeline to be marked as ready until it has successfully rebalanced.
        latch.await()
    }

    fun stop() {
        streams.close()
        // Does not clean up everything see https://issues.apache.org/jira/browse/KAFKA-13787
        streams.cleanUp()
    }

    override fun onChange(newState: KafkaStreams.State?, oldState: KafkaStreams.State?) {
        logger.info { "Pipeline ${metadata.name} (v${metadata.version}) changing to state $newState" }
        if (newState == KafkaStreams.State.RUNNING) {
            latch.countDown()
        }
    }

    companion object {
        private val logger = noCoLogger(PipelineTopology::class)
    }
}