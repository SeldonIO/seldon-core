package io.seldon.dataflow.kafka

import io.klogging.noCoLogger
import org.apache.kafka.clients.consumer.ConsumerRecord
import org.apache.kafka.clients.producer.ProducerRecord
import org.apache.kafka.streams.errors.DeserializationExceptionHandler
import org.apache.kafka.streams.errors.ProductionExceptionHandler
import org.apache.kafka.streams.errors.StreamsException
import org.apache.kafka.streams.errors.StreamsUncaughtExceptionHandler
import org.apache.kafka.streams.processor.ProcessorContext

class StreamErrorHandling {

    class StreamsDeserializationErrorHandler: DeserializationExceptionHandler {

        override fun configure(configs: MutableMap<String, *>?) {
        }

        override fun handle(
            context: ProcessorContext?,
            record: ConsumerRecord<ByteArray, ByteArray>?,
            exception: Exception?
        ): DeserializationExceptionHandler.DeserializationHandlerResponse {
            if (exception != null) {
                logger.error(exception, "Kafka streams: message deserialization error on ${record?.topic()}")
            }
            return DeserializationExceptionHandler.DeserializationHandlerResponse.CONTINUE
        }
    }

    class StreamsRecordProducerErrorHandler: ProductionExceptionHandler {

        override fun configure(configs: MutableMap<String, *>?) {
        }

        override fun handle(
            record: ProducerRecord<ByteArray, ByteArray>?,
            exception: Exception?
        ): ProductionExceptionHandler.ProductionExceptionHandlerResponse {
            if (exception != null) {
                logger.error(exception, "Kafka streams: error when writing to ${record?.topic()}")
            }
            return ProductionExceptionHandler.ProductionExceptionHandlerResponse.CONTINUE
        }

    }

    class StreamsCustomUncaughtExceptionHandler: StreamsUncaughtExceptionHandler {
        override fun handle(exception: Throwable?): StreamsUncaughtExceptionHandler.StreamThreadExceptionResponse {
            if (exception is StreamsException) {
                val originalException = exception.cause
                originalException?.let {
                    logger.error(it, "Kafka streams: stream processing exception")
                    return StreamsUncaughtExceptionHandler.StreamThreadExceptionResponse.SHUTDOWN_CLIENT;
                }
            }
            // try to continue
            return StreamsUncaughtExceptionHandler.StreamThreadExceptionResponse.REPLACE_THREAD;
        }
    }

    companion object {
        private val logger = noCoLogger(StreamErrorHandling::class)
    }

}
