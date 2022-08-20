package io.seldon.dataflow.kafka.headers

import io.seldon.dataflow.kafka.TRecord
import org.apache.kafka.streams.kstream.ValueTransformer
import org.apache.kafka.streams.processor.ProcessorContext

class PipelineNameFilter(private val pipelineName: String) : ValueTransformer<TRecord, TRecord> {
    var context: ProcessorContext? = null

    override fun init(context: ProcessorContext?) {
        this.context = context
    }

    override fun transform(value: TRecord?): TRecord? {
        val shouldProcess = context
            ?.headers()
            ?.headers(SeldonHeaders.pipelineName)
            ?.any { it.value().decodeToString() == pipelineName }
            ?: false
        return if (shouldProcess) value else null
    }

    override fun close() {}
}