package io.seldon.dataflow.kafka

import io.klogging.Klogger
import io.klogging.Level
import io.klogging.NoCoLogger
import io.seldon.dataflow.DataflowStatus
import kotlinx.coroutines.runBlocking
import org.apache.kafka.streams.KafkaStreams

open class PipelineStatus(val state: KafkaStreams.State?, var isError: Boolean) : DataflowStatus {
    // Keep the previous state in case we're stopping the stream so that we can determine
    // _why_ the stream was stopped.
    class StreamStopped(var prevState: PipelineStatus?) : PipelineStatus(null, false) {
        override var message: String? = "pipeline data streams: stopped"

        init {
            // Avoid nesting stopped states
            val prev = this.prevState
            if (prev is StreamStopped) {
                this.prevState = prev.prevState
            }
            this.isError = this.prevState?.isError ?: false
        }

        override fun getDescription() : String? {
            val exceptionMsg = this.exception?.message
            var statusMsg = this.message
            val prevStateDescription = this.prevState?.getDescription()
            prevStateDescription?.let {
                statusMsg += ", before stop: $prevStateDescription"
            }
            return if (exceptionMsg != null) {
                "$statusMsg Exception: $exceptionMsg"
            } else {
                statusMsg
            }
        }

        // log status when logger is in a coroutine
        override fun log(logger: Klogger, levelIfNoException: Level) {
            var exceptionMsg = this.exception?.message
            var exceptionCause = this.exception?.cause ?: Exception("")
            var statusMsg = this.message
            val prevStateDescription = this.prevState?.getDescription()
            prevStateDescription?.let {
                statusMsg += ", before stop: $prevStateDescription"
            }
            if (exceptionMsg != null) {
                runBlocking {
                    logger.log(levelIfNoException, exceptionCause, "$statusMsg, Exception: {exception}", exceptionMsg)
                }
            } else {
                runBlocking {
                    logger.log(levelIfNoException, "$statusMsg")
                }
            }
        }

        // log status when logger is outside coroutines
        override fun log(logger: NoCoLogger, levelIfNoException: Level) {
            val exceptionMsg = this.exception?.message
            val exceptionCause = this.exception?.cause ?: Exception("")
            var statusMsg = this.message
            val prevStateDescription = this.prevState?.getDescription()
            prevStateDescription?.let {
                statusMsg += ", stop cause: $prevStateDescription"
            }
            if (exceptionMsg != null) {
                logger.log(levelIfNoException, exceptionCause, "$statusMsg, Exception: {exception}", exceptionMsg)
            } else {
                logger.log(levelIfNoException, "$statusMsg")
            }
        }
    }

    class StreamStopping() : PipelineStatus(null, false) {
        override var message: String? = "pipeline data streams: stopping"
    }

    class StreamStarting() : PipelineStatus(null, false) {
        override var message: String? = "pipeline data streams: initializing"
    }

    class Started() : PipelineStatus(null, false) {
        override var message: String? = "pipeline data streams: ready"
    }

    data class Error(val errorState: KafkaStreams.State?): PipelineStatus(errorState, true)

    override var exception: Exception? = null
    override var message: String? = null

}
