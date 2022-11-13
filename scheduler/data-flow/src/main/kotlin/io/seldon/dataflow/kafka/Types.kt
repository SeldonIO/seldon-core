/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package io.seldon.dataflow.kafka

import org.apache.kafka.common.serialization.Serdes
import org.apache.kafka.streams.kstream.Consumed
import org.apache.kafka.streams.kstream.Produced
import org.apache.kafka.streams.kstream.StreamJoined
import java.util.*

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
    PASSTHROUGH;

    companion object {
        fun create(input: String, output: String): ChainType {
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

