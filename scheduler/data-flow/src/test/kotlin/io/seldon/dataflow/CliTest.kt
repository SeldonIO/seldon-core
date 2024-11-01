/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow

import io.seldon.dataflow.kafka.security.KafkaSaslMechanisms
import org.junit.jupiter.api.DisplayName
import org.junit.jupiter.params.ParameterizedTest
import org.junit.jupiter.params.provider.Arguments
import org.junit.jupiter.params.provider.Arguments.arguments
import org.junit.jupiter.params.provider.MethodSource
import strikt.api.expectCatching
import strikt.api.expectThat
import strikt.assertions.hasLength
import strikt.assertions.isEqualTo
import strikt.assertions.isNotEqualTo
import strikt.assertions.isSuccess
import strikt.assertions.startsWith
import java.util.UUID
import java.util.stream.Stream
import kotlin.test.Test

internal class CliTest {
    @DisplayName("Passing auth mechanism via cli argument")
    @ParameterizedTest(name = "{0}")
    @MethodSource("saslMechanisms")
    fun getSaslMechanism(
        input: String,
        expectedMechanism: KafkaSaslMechanisms,
    ) {
        val args = arrayOf("--kafka-sasl-mechanism", input)
        val cli = Cli.configWith(args)

        expectCatching { cli[Cli.saslMechanism] }
            .isSuccess()
            .isEqualTo(expectedMechanism)
    }

    @Test
    fun `should handle dataflow replica id`() {
        val cliDefault = Cli.configWith(arrayOf<String>())
        val testReplicaId = "dataflow-id-1"
        val cli = Cli.configWith(arrayOf("--dataflow-replica-id", testReplicaId))

        expectThat(cliDefault[Cli.dataflowReplicaId]) {
            isNotEqualTo("seldon-dataflow-engine")
        }
        expectThat(cli[Cli.dataflowReplicaId]) {
            isEqualTo(testReplicaId)
        }

        // test random Uuid (v4)
        val expectedReplicaIdPrefix = "seldon-dataflow-engine-"
        val uuidStringLength = 36
        val randomReplicaUuid = Cli.getNewDataflowId(true)
        expectThat(randomReplicaUuid) {
            startsWith(expectedReplicaIdPrefix)
            hasLength(expectedReplicaIdPrefix.length + uuidStringLength)
        }
        expectCatching { UUID.fromString(randomReplicaUuid.removePrefix(expectedReplicaIdPrefix)) }
            .isSuccess()
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
