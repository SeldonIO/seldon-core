/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow.kafka

import com.google.protobuf.Message
import io.confluent.kafka.schemaregistry.client.CachedSchemaRegistryClient
import io.confluent.kafka.schemaregistry.client.SchemaRegistryClient
import io.confluent.kafka.serializers.protobuf.KafkaProtobufSerializer
import io.seldon.dataflow.PipelineSubscriber
import io.seldon.mlops.inference_schema.InferRequest.ModelInferRequest
import io.seldon.mlops.inference_schema.InferResponse.ModelInferResponse
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

        val requestSerializer = KafkaProtobufSerializer<ModelInferRequest>()
        val responseSerialiser = KafkaProtobufSerializer<ModelInferResponse>()

        init {
            val props =
                mapOf(
                    "schema.registry.url" to schemaRegistryUrl,
                    "auto.register.schemas" to true,
                    "use.latest.version" to true,
                )

            requestSerializer.configure(props, false)
            responseSerialiser.configure(props, false)
        }

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
            return data?.let { removeSchemaRegistryWireFormat(topic, it) }
        }

        private fun removeSchemaRegistryWireFormat(
            topic: String?,
            data: ByteArray,
        ): ByteArray {
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

//            var dataAfter: ByteArray = byteArrayOf()
//
//            if (topic?.contains("input") ?: return data) {
//                dataAfter = deserializer.deserialize("inference_schema", data).toByteArray()
//            } else if (topic?.contains("output") ?: return data) {
//                dataAfter = deserializer.deserialize("inference_schema", data).toByteArray()
//            }

            // Skip the first 6 bytes (magic byte + 4-byte schema ID)
            val dataAfter = data.copyOfRange(6, data.size)

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
            if (data == null) return null

            return try {
                var message: Message

                if (topic?.contains("input") ?: return data) {
                    // Parse raw protobuf to your message type
                    message = ModelInferRequest.parseFrom(data)
                    val serialised = requestSerializer.serialize(topic, message)
                    logger.info(
                        "First 10 bytes after serialised of data for topic $topic: ${
                            serialised.take(10).joinToString(" ") { "%02x".format(it) }
                        }",
                    )
                    serialised
                } else if (topic?.contains("output") ?: return data) {
                    message = ModelInferResponse.parseFrom(data)

                    val serialised = responseSerialiser.serialize(topic, message)
                    logger.info(
                        "First 10 bytes after serialised of data for topic $topic: ${
                            serialised.take(10).joinToString(" ") { "%02x".format(it) }
                        }",
                    )
                    serialised
                } else {
                    data
                }
            } catch (e: Exception) {
                logger.warn("Failed to serialize with Schema Registry, using raw data", e)
                data // Fallback to raw data
            }
        }
    }

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
