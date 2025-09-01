/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow.kafka

import com.google.protobuf.Message
import io.confluent.kafka.serializers.AbstractKafkaSchemaSerDeConfig
import io.confluent.kafka.serializers.protobuf.KafkaProtobufSerializer
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
    val url: String = "",
    private val _useSchemaRegistry: Boolean? = null,
    val username: String = "",
    val password: String = "",
    val autoRegisterSchemas: Boolean = true,
    val useLatestVersion: Boolean = true,
) {
    val useSchemaRegistry: Boolean
        get() = _useSchemaRegistry ?: url.isNotBlank()

    val userInfoConfig: String
        get() =
            if (username.isNotBlank() || password.isNotBlank()) {
                "$username:$password"
            } else {
                ""
            }

    fun validate() {
        if (useSchemaRegistry) {
            require(url.isNotBlank()) { "Schema registry URL is required when useSchemaRegistry is true" }
        }
    }

    fun toSerializerProperties(): Map<String, Any> {
        return mapOf(
            AbstractKafkaSchemaSerDeConfig.SCHEMA_REGISTRY_URL_CONFIG to url,
            AbstractKafkaSchemaSerDeConfig.AUTO_REGISTER_SCHEMAS to autoRegisterSchemas,
            AbstractKafkaSchemaSerDeConfig.USE_LATEST_VERSION to useLatestVersion,
            AbstractKafkaSchemaSerDeConfig.NORMALIZE_SCHEMAS to true,
            AbstractKafkaSchemaSerDeConfig.USER_INFO_CONFIG to userInfoConfig,
            AbstractKafkaSchemaSerDeConfig.BEARER_AUTH_TOKEN_CONFIG to password,
            AbstractKafkaSchemaSerDeConfig.BASIC_AUTH_CREDENTIALS_SOURCE to "USER_INFO",
        )
    }
}

// Schema Registry Serializer Factory with proper lifecycle management
class SchemaRegistrySerializerFactory(private val config: SchemaRegistryConfig) {
    private val logger = Logger(SchemaRegistrySerializerFactory::class)

    init {
        config.validate()
        if (config.useSchemaRegistry) {
            logger.info("Initializing Schema Registry serializers with URL: ${config.url}")
        }
    }

    val requestSerializer: KafkaProtobufSerializer<ModelInferRequest> by lazy {
        createAndConfigureSerializer()
    }

    val responseSerializer: KafkaProtobufSerializer<ModelInferResponse> by lazy {
        createAndConfigureSerializer()
    }

    private fun <T : Message> createAndConfigureSerializer(): KafkaProtobufSerializer<T> {
        return try {
            val serializer = KafkaProtobufSerializer<T>()
            serializer.configure(config.toSerializerProperties(), false)

            logger.info("with user info properties ${config.toSerializerProperties()[AbstractKafkaSchemaSerDeConfig.USER_INFO_CONFIG]}")
            try {
                serializer.schemaRegistryClient.allSubjects
            } catch (e: Exception) {
                logger.info("failed to ping schema registry client")
                throw IllegalStateException("could not ping schema registry client", e)
            }

            logger.info("Successfully configured protobuf serializer")
            serializer
        } catch (e: Exception) {
            logger.error("Failed to configure protobuf serializer", e)
            throw IllegalStateException("Could not configure protobuf serializer", e)
        }
    }
}

// Wire format deserializer - extracts protobuf from Schema Registry wire format
class ProtobufWireFormatDeserializer : org.apache.kafka.common.serialization.Deserializer<ByteArray> {
    private val logger = Logger(ProtobufWireFormatDeserializer::class)

    override fun deserialize(
        topic: String?,
        data: ByteArray?,
    ): ByteArray? {
        logger.debug("Deserializing topic: $topic")
        return data?.let { removeSchemaRegistryWireFormat(topic, it) }
    }

    private fun removeSchemaRegistryWireFormat(
        topic: String?,
        data: ByteArray,
    ): ByteArray {
        // Schema Registry wire format: [magic_byte(1)] + [schema_id(5)] + [proto message index(6)] + [actual_protobuf_data...]
        logger.debug("Removing schema registry wire format from a message in $topic")

        if (data.size < 6) {
            logger.debug("No schema id in message")
            return data
        }

        // Check if first byte is the magic byte (0x0)
        if (data[0] != 0.toByte()) {
            logger.debug("Did not find magic byte, returning normal data")
            return data
        }

        logger.debug("First 10 bytes before remove: ${data.take(10).joinToString(" ") { "%02x".format(it) }}")

        // Skip the first 6 bytes (magic byte + schema ID + proto msg index)
        val dataAfter = data.copyOfRange(6, data.size)

        logger.debug("First 10 bytes after remove: ${dataAfter.take(10).joinToString(" ") { "%02x".format(it) }}")
        logger.debug("Returned data without schema id")
        return dataAfter
    }
}

// Wire format serializer - adds Schema Registry wire format to protobuf
class ProtobufWireFormatSerializer(val serializerFactory: SchemaRegistrySerializerFactory) :
    org.apache.kafka.common.serialization.Serializer<ByteArray> {
    private val logger = Logger(ProtobufWireFormatSerializer::class)

    override fun serialize(
        topic: String?,
        data: ByteArray?,
    ): ByteArray? {
        if (data == null) return null

        return try {
            when {
                topic?.contains("input") == true -> {
                    val message = ModelInferRequest.parseFrom(data)
                    val serialized = serializerFactory.requestSerializer.serialize(topic, message)
                    logger.debug("Serialized input message for topic $topic")
                    serialized
                }

                topic?.contains("output") == true -> {
                    val message = ModelInferResponse.parseFrom(data)
                    val serialized = serializerFactory.responseSerializer.serialize(topic, message)
                    logger.debug("Serialized output message for topic $topic")
                    serialized
                }

                else -> {
                    logger.debug("Topic $topic does not match input/output pattern, using raw data")
                    data
                }
            }
        } catch (e: Exception) {
            logger.warn("Failed to serialize with Schema Registry for topic $topic, using raw data", e)
            data // Fallback to raw data
        }
    }
}

// Wire format serde combining serializer and deserializer
class ProtobufWireFormatSerde(private val serializerFactory: SchemaRegistrySerializerFactory) : Serde<ByteArray> {
    override fun serializer(): org.apache.kafka.common.serialization.Serializer<ByteArray> {
        return ProtobufWireFormatSerializer(serializerFactory)
    }

    override fun deserializer(): org.apache.kafka.common.serialization.Deserializer<ByteArray> {
        return ProtobufWireFormatDeserializer()
    }
}

// Simple serde factory with single responsibility
class ProtobufSerdeFactory(private val config: SchemaRegistryConfig) {
    private val logger = Logger(ProtobufSerdeFactory::class)

    private val serializerFactory: SchemaRegistrySerializerFactory? by lazy {
        if (config.useSchemaRegistry) {
            SchemaRegistrySerializerFactory(config)
        } else {
            null
        }
    }

    fun createValueSerde(): Serde<ByteArray> {
        return when (config.useSchemaRegistry) {
            true -> {
                logger.info("Using schema registry with URL: ${config.url}")
                ProtobufWireFormatSerde(serializerFactory!!)
            }

            false -> {
                logger.info("Using raw byte array serde")
                Serdes.ByteArray()
            }
        }
    }
}

// Generic factory for Kafka Streams serdes
class KafkaSerdesFactory(config: SchemaRegistryConfig) {
    private val serdeFactory = ProtobufSerdeFactory(config)

    fun createConsumerSerde(): Consumed<String, TRecord> {
        val keySerde = Serdes.String()
        val valueSerde = serdeFactory.createValueSerde()
        return Consumed.with(keySerde, valueSerde)
    }

    fun createProducerSerde(): Produced<String, TRecord> {
        val keySerde = Serdes.String()
        val valueSerde = serdeFactory.createValueSerde()
        return Produced.with(keySerde, valueSerde, SamePartitionForwarder())
    }

    fun createJoinSerde(): StreamJoined<String, TRecord, TRecord> {
        val keySerde = Serdes.String()
        val valueSerde = Serdes.ByteArray()
        return StreamJoined.with(keySerde, valueSerde, valueSerde)
    }
}

// High-level service for Kafka Streams serdes
class KafkaStreamsSerdes(config: SchemaRegistryConfig) {
    private val serdesFactory = KafkaSerdesFactory(config)
    val consumerSerde: Consumed<String, TRecord> = serdesFactory.createConsumerSerde()
    val producerSerde: Produced<String, TRecord> = serdesFactory.createProducerSerde()
    val joinSerde: StreamJoined<String, TRecord, TRecord> = serdesFactory.createJoinSerde()
}
