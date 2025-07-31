package io.seldon.dataflow.kafka

import io.confluent.kafka.serializers.protobuf.KafkaProtobufDeserializer
import io.confluent.kafka.serializers.protobuf.KafkaProtobufSerializer
import io.confluent.kafka.streams.serdes.protobuf.KafkaProtobufSerde
import org.apache.kafka.common.serialization.Serde
import org.apache.kafka.common.serialization.Serdes
import org.apache.kafka.streams.kstream.Consumed
import org.apache.kafka.streams.kstream.Produced
import org.apache.kafka.streams.kstream.StreamJoined
import com.google.protobuf.Message

// Configuration class to hold schema registry settings
data class SchemaRegistryConfig(
    val enabled: Boolean,
    val url: String = "",
    val recordNameStrategy: String = "io.confluent.kafka.serializers.subject.RecordNameStrategy",
    val autoRegisterSchemas: Boolean = false
)

// Enhanced factory that can handle mixed environments
class ProtobufSerdeFactory {

    companion object {
        fun <T : Message> createValueSerde(
            config: SchemaRegistryConfig,
            protoClass: Class<T>
        ): Serde<T> {
            return if (config.enabled) {
                createSchemaRegistrySerde(config, protoClass)
            } else {
                createCompatibleByteArraySerde(protoClass)
            }
        }

        private fun <T : Message> createSchemaRegistrySerde(
            config: SchemaRegistryConfig,
            protoClass: Class<T>
        ): Serde<T> {
            val serde = KafkaProtobufSerde<T>()
            val props = mapOf(
                "schema.registry.url" to config.url,
                "value.subject.name.strategy" to config.recordNameStrategy,
                "auto.register.schemas" to config.autoRegisterSchemas.toString(),
                "use.latest.version" to "true"
            )
            serde.configure(props, false) // false = value serde
            return serde
        }

        // Creates a serde that can handle both raw protobuf and Schema Registry wire format
        private fun <T : Message> createCompatibleByteArraySerde(protoClass: Class<T>): Serde<T> {
            return object : Serde<T> {
                override fun serializer() = ProtobufByteArraySerializer<T>()
                override fun deserializer() = ProtobufByteArrayDeserializer(protoClass)
            }
        }

        // Creates a hybrid serde that can consume Schema Registry format but produce raw protobuf
        fun <T : Message> createHybridSerde(
            config: SchemaRegistryConfig,
            protoClass: Class<T>
        ): Serde<T> {
            return object : Serde<T> {
                override fun serializer() = ProtobufByteArraySerializer<T>()
                override fun deserializer() = if (config.enabled) {
                    // Use Schema Registry deserializer for consuming
                    val deserializer = KafkaProtobufDeserializer<T>()
                    val props = mapOf(
                        "schema.registry.url" to config.url,
                        "value.subject.name.strategy" to config.recordNameStrategy,
                        "use.latest.version" to "true"
                    )
                    deserializer.configure(props, false)
                    deserializer
                } else {
                    ProtobufByteArrayDeserializer(protoClass)
                }
            }
        }
    }
}

// Custom serializer that handles both raw protobuf and Schema Registry wire format
class ProtobufByteArraySerializer<T : Message> : org.apache.kafka.common.serialization.Serializer<T> {
    override fun serialize(topic: String?, data: T?): ByteArray? {
        return data?.toByteArray()
    }
}

// Custom deserializer that can handle both raw protobuf and Schema Registry wire format
class ProtobufByteArrayDeserializer<T : Message>(
    private val protoClass: Class<T>
) : org.apache.kafka.common.serialization.Deserializer<T> {

    companion object {
        private const val MAGIC_BYTE: Byte = 0x0
        private const val SCHEMA_ID_SIZE = 4
        private const val WIRE_FORMAT_HEADER_SIZE = 5 // magic byte + 4 bytes schema ID
    }

    override fun deserialize(topic: String?, data: ByteArray?): T? {
        if (data == null) return null

        val protobufData = if (isSchemaRegistryWireFormat(data)) {
            // Skip the Schema Registry wire format header (magic byte + schema ID)
            data.copyOfRange(WIRE_FORMAT_HEADER_SIZE, data.size)
        } else {
            // Raw protobuf data
            data
        }

        return parseProtobufMessage(protobufData)
    }

    private fun isSchemaRegistryWireFormat(data: ByteArray): Boolean {
        return data.size > WIRE_FORMAT_HEADER_SIZE && data[0] == MAGIC_BYTE
    }

    private fun parseProtobufMessage(data: ByteArray): T? {
        return try {
            // For Kotlin generated protobuf classes, parseFrom is in the companion object
            val companionClass = Class.forName("${protoClass.name}\$Companion")
            val companionField = protoClass.getDeclaredField("Companion")
            val companion = companionField.get(null)

            val parseMethod = companionClass.getMethod("parseFrom", ByteArray::class.java)
            parseMethod.invoke(companion, data) as T
        } catch (e: Exception) {
            // Fallback to Java-style if Kotlin companion object approach fails
            try {
                val parseMethod = protoClass.getMethod("parseFrom", ByteArray::class.java)
                parseMethod.invoke(null, data) as T
            } catch (fallbackException: Exception) {
                throw RuntimeException("Failed to deserialize protobuf message. Tried both Kotlin and Java approaches.", e)
            }
        }
    }
}

// Factory class to create your specific serdes
class KafkaSerdesFactory<TRecord : Message>(
    private val config: SchemaRegistryConfig,
    private val recordClass: Class<TRecord>
) {

    fun createConsumerSerde(): Consumed<String, TRecord> {
        val keySerde = Serdes.String()
        val valueSerde = ProtobufSerdeFactory.createValueSerde(config, recordClass)
        return Consumed.with(keySerde, valueSerde)
    }

    fun createProducerSerde(): Produced<String, TRecord> {
        val keySerde = Serdes.String()
        val valueSerde = ProtobufSerdeFactory.createValueSerde(config, recordClass)
        return Produced.with(keySerde, valueSerde, SamePartitionForwarder())
    }

    fun createJoinSerde(): StreamJoined<String, TRecord, TRecord> {
        val keySerde = Serdes.String()
        val valueSerde = ProtobufSerdeFactory.createValueSerde(config, recordClass)
        return StreamJoined.with(keySerde, valueSerde, valueSerde)
    }
}

// Usage example with Kotlin protobuf classes
class KafkaStreamsService<TRecord : Message>(
    private val recordClass: Class<TRecord>,
    schemaRegistryUrl: String,
    useSchemaRegistry: Boolean
) {

    private val config = SchemaRegistryConfig(
        enabled = useSchemaRegistry,
        url = schemaRegistryUrl,
        recordNameStrategy = "io.confluent.kafka.serializers.subject.RecordNameStrategy",
        autoRegisterSchemas = false
    )

    private val serdesFactory = KafkaSerdesFactory(config, recordClass)

    // Your serdes - now configurable
    val consumerSerde: Consumed<String, TRecord> = serdesFactory.createConsumerSerde()
    val producerSerde: Produced<String, TRecord> = serdesFactory.createProducerSerde()
    val joinSerde: StreamJoined<String, TRecord, TRecord> = serdesFactory.createJoinSerde()

    // Example of how to use in your topology
    fun buildTopology(): org.apache.kafka.streams.Topology {
        val builder = org.apache.kafka.streams.StreamsBuilder()

        val stream = builder.stream<String, TRecord>("input-topic", consumerSerde)

        stream
            .mapValues { value ->
                // Your processing logic here
                value
            }
            .to("output-topic", producerSerde)

        return builder.build()
    }
}

// Example usage with Kotlin protobuf classes
// Assuming you have a Kotlin-generated protobuf class like MyProtoRecordKt
/*
val kafkaService = KafkaStreamsService(
    recordClass = MyProtoRecordKt::class.java,
    schemaRegistryUrl = "http://localhost:8081",
    useSchemaRegistry = true
)

// Alternative: Create serdes directly for inline usage
val config = SchemaRegistryConfig(
    enabled = true,
    url = "http://localhost:8081"
)

val myConsumerSerde: Consumed<String, MyProtoRecordKt> = Consumed.with(
    Serdes.String(),
    ProtobufSerdeFactory.createValueSerde(config, MyProtoRecordKt::class.java)
)
*/

// Configuration loading example
object ConfigLoader {
    fun loadSchemaRegistryConfig(): SchemaRegistryConfig {
        // Load from your configuration source (application.yml, environment variables, etc.)
        val useSchemaRegistry = System.getenv("USE_SCHEMA_REGISTRY")?.toBoolean() ?: false
        val schemaRegistryUrl = System.getenv("SCHEMA_REGISTRY_URL") ?: "http://localhost:8081"

        return SchemaRegistryConfig(
            enabled = useSchemaRegistry,
            url = schemaRegistryUrl,
            recordNameStrategy = "io.confluent.kafka.serializers.subject.RecordNameStrategy",
            autoRegisterSchemas = false
        )
    }
}