package io.seldon.dataflow.kafka.headers

import io.seldon.dataflow.kafka.TRecord
import org.apache.kafka.streams.kstream.ValueTransformer
import org.apache.kafka.streams.processor.ProcessorContext

class PipelineHeaderSetter(private val pipelineName: String) : ValueTransformer<TRecord, TRecord> {
    var context: ProcessorContext? = null

    override fun init(context: ProcessorContext?) {
        this.context = context
    }

    override fun transform(value: TRecord?): TRecord? {
        this.context?.headers()?.remove(SeldonHeaders.pipelineName)
        this.context?.headers()?.add(SeldonHeaders.pipelineName, pipelineName.toByteArray())
        return value
    }

    override fun close() {}
}