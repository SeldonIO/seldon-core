/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow.kafka

import org.apache.kafka.common.serialization.Serde
import org.apache.kafka.common.serialization.Serdes
import org.apache.kafka.streams.kstream.Consumed
import org.apache.kafka.streams.kstream.Produced
import org.apache.kafka.streams.kstream.StreamJoined

// Configuration class to hold schema registry settings
data class SchemaRegistryConfig(
    val enabled: Boolean,
    val url: String = "",
    val recordNameStrategy: String = "io.confluent.kafka.serializers.subject.RecordNameStrategy",
    val autoRegisterSchemas: Boolean = false,
    val consumeSchemaRegistryFormat: Boolean = enabled,
)

// Enhanced factory with clear configuration-based choices
class ProtobufSerdeFactory {
    companion object {
        fun createValueSerde(useSchemaRegistry: Boolean): Serde<ByteArray> {
            return when (useSchemaRegistry) {
                true -> {
                    println("Using schema registry")
                    createWireFormatByteArraySerde()
                }

                false -> createRawByteArraySerde()
            }
        }

//        private fun createSchemaRegistrySerde(
//            config: SchemaRegistryConfig,
//        ): Serde<ByteArray> {
//            val serde = KafkaProtobufSerde<ByteArray>()
//            val props = mapOf(
//                "schema.registry.url" to config.url,
//                "value.subject.name.strategy" to config.recordNameStrategy,
//                "auto.register.schemas" to config.autoRegisterSchemas.toString(),
//                "use.latest.version" to "true"
//            )
//            serde.configure(props, false) // false = value serde
//            return serde.toString().toByteArray()
//        }

        // For consuming messages with Schema Registry wire format but without schema validation
        private fun createWireFormatByteArraySerde(): Serde<ByteArray> {
//            return object : Serde<ByteArray> {
//                override fun serializer() = ProtobufByteArraySerializer<T>()
//                override fun deserializer() = SchemaRegistryWireFormatDeserializer()
//            }
            return Serdes.ByteArray()
        }

        // For consuming raw protobuf messages like before
        private fun createRawByteArraySerde(): Serde<ByteArray> {
            return Serdes.ByteArray()
        }
    }

    // Raw protobuf deserializer (no wire format handling)
//    class ProtobufByteArrayDeserializer<T : Message>(
//        private val protoClass: Class<T>
//    ) : org.apache.kafka.common.serialization.Deserializer<T> {
//
//        override fun deserialize(topic: String?, data: ByteArray?): T? {
//            if (data == null) return null
//
//            return try {
//                // For Kotlin generated protobuf classes, parseFrom is in the companion object
//                val companionClass = Class.forName("${protoClass.name}\$Companion")
//                val companionField = protoClass.getDeclaredField("Companion")
//                val companion = companionField.get(null)
//
//                val parseMethod = companionClass.getMethod("parseFrom", ByteArray::class.java)
//                parseMethod.invoke(companion, data) as T
//            } catch (e: Exception) {
//                // Fallback to Java-style if Kotlin companion object approach fails
//                try {
//                    val parseMethod = protoClass.getMethod("parseFrom", ByteArray::class.java)
//                    parseMethod.invoke(null, data) as T
//                } catch (fallbackException: Exception) {
//                    throw RuntimeException(
//                        "Failed to deserialize protobuf message. Tried both Kotlin and Java approaches.",
//                        e
//                    )
//                }
//            }
//        }
//    }

    // Wire format deserializer (strips Schema Registry header then deserializes protobuf)
//    class SchemaRegistryWireFormatDeserializer<T : Message>(
//        private val protoClass: Class<T>
//    ) : org.apache.kafka.common.serialization.Deserializer<T> {
//
//        companion object {
//            private const val WIRE_FORMAT_HEADER_SIZE = 5 // magic byte + 4 bytes schema ID
//        }
//
//        override fun deserialize(topic: String?, data: ByteArray?): T? {
//            if (data == null) return null
//
//            // Strip the Schema Registry wire format header (magic byte + schema ID)
//            val protobufData = if (data.size > WIRE_FORMAT_HEADER_SIZE) {
//                data.copyOfRange(WIRE_FORMAT_HEADER_SIZE, data.size)
//            } else {
//                throw RuntimeException("Invalid Schema Registry wire format: data too short")
//            }
//
//            return parseProtobufMessage(protobufData)
//        }
//
//        private fun parseProtobufMessage(data: ByteArray): T? {
//            return try {
//                // For Kotlin generated protobuf classes, parseFrom is in the companion object
//                val companionClass = Class.forName("${protoClass.name}\$Companion")
//                val companionField = protoClass.getDeclaredField("Companion")
//                val companion = companionField.get(null)
//
//                val parseMethod = companionClass.getMethod("parseFrom", ByteArray::class.java)
//                parseMethod.invoke(companion, data) as T
//            } catch (e: Exception) {
//                // Fallback to Java-style if Kotlin companion object approach fails
//                try {
//                    val parseMethod = protoClass.getMethod("parseFrom", ByteArray::class.java)
//                    parseMethod.invoke(null, data) as T
//                } catch (fallbackException: Exception) {
//                    throw RuntimeException(
//                        "Failed to deserialize protobuf message. Tried both Kotlin and Java approaches.",
//                        e
//                    )
//                }
//            }
//        }
//    }

    // Generic factory for any protobuf class
    class KafkaSerdesFactory(private val useSchemaRegistry: Boolean) {
        fun createConsumerSerde(): Consumed<String, TRecord> {
            val keySerde = Serdes.String()
            val valueSerde = createValueSerde(useSchemaRegistry)
            return Consumed.with(keySerde, valueSerde)
        }

        fun createProducerSerde(): Produced<String, TRecord> {
            val keySerde = Serdes.String()
            val valueSerde = createValueSerde(useSchemaRegistry)
            return Produced.with(keySerde, valueSerde, SamePartitionForwarder())
        }

        fun createJoinSerde(): StreamJoined<String, TRecord, TRecord> {
            val keySerde = Serdes.String()
            val valueSerde = Serdes.ByteArray()
            return StreamJoined.with(keySerde, valueSerde, valueSerde)
        }
    }

    // Serdes provider service - focused only on providing configured serdes
    class KafkaStreamsSerdes(private val useSchemaRegistry: Boolean) {
        private val serdesFactory = KafkaSerdesFactory(useSchemaRegistry)
        val consumerSerde: Consumed<String, TRecord> = serdesFactory.createConsumerSerde()
        val producerSerde: Produced<String, TRecord> = serdesFactory.createProducerSerde()
        val joinSerde: StreamJoined<String, TRecord, TRecord> = serdesFactory.createJoinSerde()
    }

    // Configuration loading example
    object ConfigLoader {
        fun loadSchemaRegistryConfig(): SchemaRegistryConfig {
            val useSchemaRegistry = System.getenv("USE_SCHEMA_REGISTRY")?.toBoolean() ?: false
            val consumeSchemaRegistryFormat =
                System.getenv("CONSUME_SCHEMA_REGISTRY_FORMAT")?.toBoolean() ?: useSchemaRegistry
            val schemaRegistryUrl = System.getenv("SCHEMA_REGISTRY_URL") ?: "http://localhost:8081"

            return SchemaRegistryConfig(
                enabled = useSchemaRegistry,
                url = schemaRegistryUrl,
                recordNameStrategy = "io.confluent.kafka.serializers.subject.RecordNameStrategy",
                autoRegisterSchemas = false,
                consumeSchemaRegistryFormat = consumeSchemaRegistryFormat,
            )
        }
    }
}
