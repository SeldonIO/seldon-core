package io.seldon.dataflow

import io.seldon.dataflow.kafka.security.KafkaSaslMechanisms
import org.junit.jupiter.api.Test

import org.junit.jupiter.api.Assertions.*

internal class CliTest {

    @Test
    fun getSaslMechanism() {
        val expectedMechanism = KafkaSaslMechanisms.SCRAM_SHA_512
        val args = arrayOf("--kafka-sasl-mechanism", "SCRAM-SHA-512")
        val cli = Cli.configWith(args)

        val actualMechanism = cli[Cli.saslMechanism]

        assertEquals(expectedMechanism, actualMechanism)
    }
}