package io.seldon.dataflow

import io.seldon.dataflow.kafka.security.KafkaSaslMechanisms
import org.junit.jupiter.api.Test
import strikt.api.expectCatching
import strikt.assertions.isEqualTo
import strikt.assertions.isSuccess

internal class CliTest {

    @Test
    fun getSaslMechanism() {
        val expectedMechanism = KafkaSaslMechanisms.SCRAM_SHA_512
        val args = arrayOf("--kafka-sasl-mechanism", "SCRAM-SHA-512")
        val cli = Cli.configWith(args)

        expectCatching { cli[Cli.saslMechanism] }
            .isSuccess()
            .isEqualTo(expectedMechanism)
    }

        assertEquals(expectedMechanism, actualMechanism)
    }
}