package io.seldon.dataflow

import io.seldon.dataflow.kafka.security.KafkaSaslMechanisms
import org.junit.jupiter.params.ParameterizedTest
import org.junit.jupiter.params.provider.Arguments
import org.junit.jupiter.params.provider.Arguments.arguments
import org.junit.jupiter.params.provider.MethodSource
import strikt.api.expectCatching
import strikt.assertions.isEqualTo
import strikt.assertions.isSuccess
import java.util.stream.Stream

internal class CliTest {

    @ParameterizedTest(name = "{0}")
    @MethodSource("saslMechanisms")
    fun getSaslMechanism(input: String, expectedMechanism: KafkaSaslMechanisms) {
        val args = arrayOf("--kafka-sasl-mechanism", input)
        val cli = Cli.configWith(args)

        expectCatching { cli[Cli.saslMechanism] }
            .isSuccess()
            .isEqualTo(expectedMechanism)
    }

    companion object {
        @JvmStatic
        private fun saslMechanisms(): Stream<Arguments> {
            return Stream.of(
                arguments("SCRAM-SHA-512", KafkaSaslMechanisms.SCRAM_SHA_512),
                arguments("SCRAM-SHA-256", KafkaSaslMechanisms.SCRAM_SHA_256),
                arguments("PLAIN", KafkaSaslMechanisms.PLAIN),
            )
        }
    }
}