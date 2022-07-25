package io.seldon.dataflow.kafka

import io.klogging.noCoLogger
import org.apache.kafka.streams.kstream.ValueTransformer
import org.apache.kafka.streams.processor.ProcessorContext

class HeaderRemover() : ValueTransformer<TRecord, TRecord> {
    var context: ProcessorContext? = null
    val headers = arrayOf(
        "x-seldon-alibi-type",
        "x-seldon-alibi-method")

    override fun init(context: ProcessorContext?) {
        this.context = context
    }

    // Remove headers we don't want transferred between topics
    // These headers are set in MLServer Alibi-detect and Alibi-explain runtimes: https://github.com/SeldonIO/mlserver
    override fun transform(value: TRecord?): TRecord? {
        this.headers.map { header -> this.context?.headers()?.remove(header) }
        return value
    }

    override fun close() {
    }

    companion object {
        private val logger = noCoLogger(HeaderFilter::class)
    }
}