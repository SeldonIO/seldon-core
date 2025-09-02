package io.seldon.dataflow.kafka

import io.confluent.kafka.serializers.AbstractKafkaSchemaSerDeConfig
import io.seldon.mlops.inference_schema.InferResponse.ModelInferResponse
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Nested
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import strikt.api.expectThat
import strikt.assertions.contains
import strikt.assertions.hasSize
import strikt.assertions.isEqualTo
import strikt.assertions.isFalse
import strikt.assertions.isNotNull
import strikt.assertions.isNull
import strikt.assertions.isTrue

class SchemaRegistryTest {
    @Nested
    inner class SchemaRegistryConfigTest {
        @Test
        fun `validate passes with valid config`() {
            val config =
                SchemaRegistryConfig(
                    url = "mock://",
                )

            // Should not throw
            config.validate()
        }

        @Test
        fun `validate fails when useSchemaRegistry true but url blank`() {
            val config =
                SchemaRegistryConfig(
                    url = "",
                    _useSchemaRegistry = true,
                )

            val exception = assertThrows<IllegalArgumentException> { config.validate() }
            expectThat(exception.message).isNotNull().contains("Schema registry URL is required")
        }

        @Test
        fun `validate password is set when user is empty`() {
            val config =
                SchemaRegistryConfig(
                    url = "mock://",
                    password = "test-pass",
                )

            expectThat(config.userInfoConfig).isEqualTo(":test-pass")
        }

        @Test
        fun `validate user name is set when password is empty`() {
            val config =
                SchemaRegistryConfig(
                    url = "mock://",
                    username = "test-user",
                )

            expectThat(config.userInfoConfig).isEqualTo("test-user:")
        }

        @Test
        fun `useSchemaRegistry auto-detects from non-blank url`() {
            val config = SchemaRegistryConfig(url = "mock://")

            expectThat(config.useSchemaRegistry).isTrue()
        }

        @Test
        fun `useSchemaRegistry auto-detects false from blank url`() {
            val config = SchemaRegistryConfig(url = "")

            expectThat(config.useSchemaRegistry).isFalse()
        }

        @Test
        fun `useSchemaRegistry respects explicit false even with url`() {
            val config =
                SchemaRegistryConfig(
                    url = "mock://",
                    _useSchemaRegistry = false,
                )

            expectThat(config.useSchemaRegistry).isFalse()
        }

        @Test
        fun `toSerializerProperties contains all required properties`() {
            val config =
                SchemaRegistryConfig(
                    url = "mock://",
                    autoRegisterSchemas = true,
                    useLatestVersion = false,
                )

            val properties = config.toSerializerProperties()

            expectThat(properties).hasSize(7)
            expectThat(properties[AbstractKafkaSchemaSerDeConfig.SCHEMA_REGISTRY_URL_CONFIG]).isEqualTo("mock://")
            expectThat(properties[AbstractKafkaSchemaSerDeConfig.AUTO_REGISTER_SCHEMAS]).isEqualTo(true)
            expectThat(properties[AbstractKafkaSchemaSerDeConfig.USE_LATEST_VERSION]).isEqualTo(false)
            expectThat(properties[AbstractKafkaSchemaSerDeConfig.NORMALIZE_SCHEMAS]).isEqualTo(true)
            expectThat(properties[AbstractKafkaSchemaSerDeConfig.USER_INFO_CONFIG]).isEqualTo("")
            expectThat(properties[AbstractKafkaSchemaSerDeConfig.BEARER_AUTH_TOKEN_CONFIG]).isEqualTo("")
            expectThat(properties[AbstractKafkaSchemaSerDeConfig.BASIC_AUTH_CREDENTIALS_SOURCE]).isEqualTo("USER_INFO")
        }
    }

    @Nested
    inner class ProtobufWireFormatDeserializerTest {
        private lateinit var deserializer: ProtobufWireFormatDeserializer

        @BeforeEach
        fun setup() {
            deserializer = ProtobufWireFormatDeserializer()
        }

        @Test
        fun `deserialize returns null for null input`() {
            val result = deserializer.deserialize("test-topic", null)

            expectThat(result).isNull()
        }

        @Test
        fun `deserialize returns original data when too small for wire format`() {
            val data = byteArrayOf(0x01, 0x02, 0x03)

            val result = deserializer.deserialize("test-topic", data)

            expectThat(result).isEqualTo(data)
        }

        @Test
        fun `deserialize returns original data when magic byte missing`() {
            val data = byteArrayOf(0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07)

            val result = deserializer.deserialize("test-topic", data)

            expectThat(result).isEqualTo(data)
        }

        @Test
        fun `deserialize removes wire format correctly with magic byte`() {
            val protobufData = byteArrayOf(0x08, 0x01, 0x12, 0x04, 0x74, 0x65, 0x73, 0x74)
            val wireFormatData = byteArrayOf(0x00, 0x01, 0x02, 0x03, 0x04, 0x05) + protobufData

            val result = deserializer.deserialize("test-topic", wireFormatData)

            expectThat(result).isEqualTo(protobufData)
        }

        @Test
        fun `deserialize handles minimum valid wire format size`() {
            val protobufData = byteArrayOf(0x08)
            val wireFormatData = byteArrayOf(0x00, 0x01, 0x02, 0x03, 0x04, 0x05) + protobufData

            val result = deserializer.deserialize("test-topic", wireFormatData)

            expectThat(result).isEqualTo(protobufData)
        }

        @Test
        fun `deserialize returns same data if if does not include valid wire format`() {
            val wireFormatData =
                ModelInferResponse.newBuilder()
                    .setModelName("test-model")
                    .setId("test-id")
                    .build().toByteArray()

            val result = deserializer.deserialize("test-topic", wireFormatData)

            expectThat(result).isEqualTo(wireFormatData)
        }
    }

    @Nested
    inner class SchemaRegistrySerializerFactoryTest {
        @Test
        fun `constructor validates config`() {
            val validConfig = SchemaRegistryConfig(url = "mock://")

            // Should not throw
            SchemaRegistrySerializerFactory(validConfig)
        }

        @Test
        fun `responseSerializer lazy initialization works`() {
            val config = SchemaRegistryConfig(url = "mock://")
            val factory = SchemaRegistrySerializerFactory(config)

            expectThat(factory.responseSerializer)
        }

        @Test
        fun `responseSerializer serialising data`() {
            val deserializer = ProtobufWireFormatDeserializer()
            val config = SchemaRegistryConfig(url = "mock://")
            val factory = SchemaRegistrySerializerFactory(config)

            val testResponse =
                ModelInferResponse.newBuilder()
                    .setModelName("test-model")
                    .setId("test-id")
                    .build()

            val result = factory.responseSerializer.serialize("test-topic", testResponse)

            expectThat(result).isNotNull()

            val responseBytes = deserializer.deserialize("test-topic", result)
            val responseAfterDeserialisation = ModelInferResponse.parseFrom(responseBytes)

            expectThat(testResponse).equals(responseAfterDeserialisation)
        }
    }
}
