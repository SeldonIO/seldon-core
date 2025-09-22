/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow.kafka.headers

import io.seldon.dataflow.kafka.TRecord
import org.apache.kafka.streams.processor.api.FixedKeyProcessor
import org.apache.kafka.streams.processor.api.FixedKeyProcessorContext
import org.apache.kafka.streams.processor.api.FixedKeyRecord

class PipelineNameFilter<T>(private val pipelineName: String) : FixedKeyProcessor<T, TRecord, TRecord> {
    private var context: FixedKeyProcessorContext<T, TRecord>? = null

    override fun init(context: FixedKeyProcessorContext<T, TRecord>?) {
        this.context = context
    }

    override fun process(record: FixedKeyRecord<T, TRecord>?) {
        val shouldProcess =
            record
                ?.headers()
                ?.headers(SeldonHeaders.PIPELINE_NAME)
                ?.any { it.value().decodeToString() == pipelineName }
                ?: false
        if (shouldProcess) context?.forward(record)
    }

    override fun close() = Unit
}
