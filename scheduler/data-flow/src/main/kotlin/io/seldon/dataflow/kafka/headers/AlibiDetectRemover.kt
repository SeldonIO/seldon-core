package io.seldon.dataflow.kafka.headers

import io.seldon.dataflow.kafka.TRecord
import org.apache.kafka.streams.kstream.ValueTransformer
import org.apache.kafka.streams.processor.ProcessorContext

class AlibiDetectRemover : ValueTransformer<TRecord, TRecord> {
    var context: ProcessorContext? = null

    override fun init(context: ProcessorContext?) {
        this.context = context
    }

    override fun transform(value: TRecord?): TRecord? {
        SeldonHeaders.alibiDiscards
            .forEach { header ->
                this.context?.headers()?.remove(header)
            }
        return value
    }

    override fun close() {}
}