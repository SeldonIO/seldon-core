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

class PipelineHeaderProcessor<T>() : FixedKeyProcessor<T, TRecord, TRecord> {
    private var context: FixedKeyProcessorContext<T, TRecord>? = null
    private val headersToAdd: MutableMap<String, String> = mutableMapOf()
    private val headersToRemove: MutableList<String> = mutableListOf()

    fun addHeader(
        key: String,
        value: String,
    ): PipelineHeaderProcessor<T> {
        this.headersToAdd.put(key, value)
        return this
    }

    fun removeHeader(key: String): PipelineHeaderProcessor<T> {
        this.headersToRemove.add(key)
        return this
    }

    override fun init(context: FixedKeyProcessorContext<T, TRecord>?) {
        this.context = context
    }

    override fun process(record: FixedKeyRecord<T, TRecord>?) {
        val toFwd = record?.withHeaders(record.headers())
        // first remove the headers we were asked to, in case some of the added headers
        // overwrite existing header values
        this.headersToRemove.forEach { header ->
            toFwd?.headers()?.remove(header)
        }
        this.headersToAdd.forEach { (header, value) ->
            toFwd?.headers()?.add(header, value.toByteArray())
        }

        this.context?.forward(toFwd)
    }

    override fun close() = Unit
}
