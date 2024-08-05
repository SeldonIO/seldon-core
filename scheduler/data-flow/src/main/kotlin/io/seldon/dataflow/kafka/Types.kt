/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow.kafka

import org.apache.kafka.common.serialization.Serdes
import org.apache.kafka.streams.kstream.Consumed
import org.apache.kafka.streams.kstream.Produced
import org.apache.kafka.streams.kstream.StreamJoined
import java.util.Properties

typealias KafkaProperties = Properties
typealias KafkaAdminProperties = Properties
typealias TopicName = String
typealias TensorName = String
typealias RequestId = String
typealias TRecord = ByteArray
typealias TopicsAndTensors = Pair<Set<TopicName>, Set<TensorName>>

val consumerSerde: Consumed<RequestId, TRecord> = Consumed.with(Serdes.String(), Serdes.ByteArray())
val producerSerde: Produced<RequestId, TRecord> = Produced.with(Serdes.String(), Serdes.ByteArray())
val joinSerde: StreamJoined<RequestId, TRecord, TRecord> =
    StreamJoined.with(Serdes.String(), Serdes.ByteArray(), Serdes.ByteArray())

enum class ChainType {
    OUTPUT_OUTPUT,
    INPUT_INPUT,
    INPUT_OUTPUT,
    OUTPUT_INPUT,
    PASSTHROUGH,
    ;

    companion object {
        fun create(
            input: String,
            output: String,
        ): ChainType {
            return when (input.substringAfterLast(".") to output.substringAfterLast(".")) {
                "inputs" to "inputs" -> INPUT_INPUT
                "inputs" to "outputs" -> INPUT_OUTPUT
                "outputs" to "outputs" -> OUTPUT_OUTPUT
                "outputs" to "inputs" -> OUTPUT_INPUT
                else -> PASSTHROUGH
            }
        }
    }
}
