/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/


package io.seldon.dataflow.kafka.headers

import io.seldon.dataflow.kafka.TRecord
import org.apache.kafka.streams.kstream.ValueTransformer
import org.apache.kafka.streams.processor.ProcessorContext

class PipelineHeaderSetter(private val pipelineName: String, private val pipelineVersion: String) : ValueTransformer<TRecord, TRecord> {
    var context: ProcessorContext? = null

    override fun init(context: ProcessorContext?) {
        this.context = context
    }

    override fun transform(value: TRecord?): TRecord? {
        this.context?.headers()?.remove(SeldonHeaders.pipelineName)
        this.context?.headers()?.add(SeldonHeaders.pipelineName, pipelineName.toByteArray())
        this.context?.headers()?.remove(SeldonHeaders.pipelineVersion)
        this.context?.headers()?.add(SeldonHeaders.pipelineVersion, pipelineVersion.toByteArray())
        return value
    }

    override fun close() {}
}