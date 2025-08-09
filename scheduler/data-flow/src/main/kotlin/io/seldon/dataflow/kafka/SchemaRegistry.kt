/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow.kafka

import io.confluent.kafka.schemaregistry.client.CachedSchemaRegistryClient
import io.confluent.kafka.schemaregistry.client.SchemaRegistryClient
import io.seldon.dataflow.PipelineSubscriber
import org.apache.kafka.common.serialization.Serde
import org.apache.kafka.common.serialization.Serdes
import org.apache.kafka.streams.kstream.Consumed
import org.apache.kafka.streams.kstream.Produced
import org.apache.kafka.streams.kstream.StreamJoined
import io.klogging.noCoLogger as Logger

// Configuration class to hold schema registry settings
data class SchemaRegistryConfig(
    val enabled: Boolean,
    val url: String = "",
    val recordNameStrategy: String = "io.confluent.kafka.serializers.subject.RecordNameStrategy",
    val autoRegisterSchemas: Boolean = false,
    val consumeSchemaRegistryFormat: Boolean = enabled,
)

// Enhanced factory with clear configuration-based choices
class SerdeFactory {
    companion object {
        fun createValueSerde(useSchemaRegistry: Boolean): Serde<ByteArray> {
            return when (useSchemaRegistry) {
                true -> {
                    logger.info("Using schema registry")
                    createWireFormatByteArraySerde()
                }

                false -> createRawByteArraySerde()
            }
        }

        val schemaRegistryUrl = "http://schema-registry.confluent.svc.cluster.local:8081"
        val basicAuthCredentialsSource = ""
        val basicAuthUserInfo = ""

        val configs =
            mapOf(
                "basic.auth.credentials.source" to basicAuthCredentialsSource,
                "basic.auth.user.info" to basicAuthUserInfo,
            )

        val schemaClient: SchemaRegistryClient =
            CachedSchemaRegistryClient(
                schemaRegistryUrl,
                100,
                configs,
            )

//        val deserializer = KafkaProtobufDeserializer<DynamicMessage>(schemaClient)
//
//        init {
//            val props =
//                mapOf(
//                    "schema.registry.url" to schemaRegistryUrl,
//                    "specific.protobuf.reader" to false,
//                    // Use DynamicMessage instead of specific classes
//                )
//            deserializer.configure(props, false)
//        }

        private val logger = Logger(PipelineSubscriber::class)

        // For consuming messages with Schema Registry wire format but without schema validation
        private fun createWireFormatByteArraySerde(): Serde<ByteArray> {
            return object : Serde<ByteArray> {
                override fun serializer() = ProtobufByteArraySerializer()

                override fun deserializer() = ProtobufByteArrayDeserializer()
            }
        }

        // For consuming raw protobuf messages like before
        private fun createRawByteArraySerde(): Serde<ByteArray> {
            return Serdes.ByteArray()
        }
    }

    class ProtobufByteArrayDeserializer() : org.apache.kafka.common.serialization.Deserializer<ByteArray> {
        override fun deserialize(
            topic: String?,
            data: ByteArray?,
        ): ByteArray? {
            logger.info("the topic to deserialize is $topic")
            return data?.let { removeSchemaRegistryWireFormat(it) }
        }

        private fun removeSchemaRegistryWireFormat(data: ByteArray): ByteArray {
            // Schema Registry wire format:
            // [magic_byte(1)] + [schema_id(4)] + [actual_protobuf_data...]

            logger.info("Removing schema registry")
            if (data.size < 5) {
                logger.info("No schema id in message")
                // Not enough bytes for wire format, return as-is
                return data
            }

            // Check if first byte is the magic byte (0x0)
            if (data[0] != 0.toByte()) {
                logger.info("did not find magic byte, returning normal data")
                // Not schema registry format, return as-is
                return data
            }

            logger.info("First 10 bytes before remove: ${data.take(10).joinToString(" ") { "%02x".format(it) }}")

            // Skip the first 5 bytes (magic byte + 4-byte schema ID)
            val dataAfter = data.copyOfRange(7, data.size)

            logger.info("First 10 bytes after remove: ${dataAfter.take(10).joinToString(" ") { "%02x".format(it) }}")
            logger.info("Returned data without schema id")
            return dataAfter
        }
    }

    class ProtobufByteArraySerializer : org.apache.kafka.common.serialization.Serializer<ByteArray> {
        override fun serialize(
            topic: String?,
            data: ByteArray?,
        ): ByteArray? {
            logger.info("the topic to serialise data is $topic")
            return data?.let { payload ->
                val schemaId = getSchemaId()
                if (schemaId != null) {
                    addSchemaRegistryWireFormat(payload, schemaId)
                } else {
                    logger.info("Schema with id $schemaId not found")
                    // No schema ID available, return original payload without wire format
                    payload
                }
            }
        }

        private fun addSchemaRegistryWireFormat(
            data: ByteArray,
            schemaId: Int,
        ): ByteArray {
            logger.info("Adding schema registry with id $schemaId")

            logger.info("First 10 bytes before adding schema: ${data.take(10).joinToString(" ") { "%02x".format(it) }}")
            val result = ByteArray(5 + data.size)
            result[0] = 0 // Magic byte

            // Write schema ID as 4 bytes (big-endian)
            result[1] = (schemaId shr 24).toByte()
            result[2] = (schemaId shr 16).toByte()
            result[3] = (schemaId shr 8).toByte()
            result[4] = schemaId.toByte()

            // Copy the actual protobuf data
            data.copyInto(result, 5)

            logger.info(
                "First 10 bytes after adding schema: ${
                    result.take(10).joinToString(" ") { "%02x".format(it) }
                }",
            )
            return result
        }

        private fun getSchemaId(subject: String = "infer_response"): Int? {
            return try {
                val metadata = schemaClient.getLatestSchemaMetadata(subject)
                metadata.id
            } catch (e: Exception) {
                logger.warn("Could not get schema metadata for subject $subject", e)
                null // Return null to indicate failure
            }
        }
    }

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
    class KafkaStreamsSerdes(useSchemaRegistry: Boolean) {
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
