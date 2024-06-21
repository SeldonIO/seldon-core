/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow.kafka.security

sealed class SaslConfig(
    val mechanism: KafkaSaslMechanisms,
) {
    class Oauth(
        val secretName: String,
    ) : SaslConfig(KafkaSaslMechanisms.OAUTH_BEARER)

    sealed class Password(
        mechanism: KafkaSaslMechanisms,
        val secretName: String,
        val username: String,
        val passwordField: String,
    ) : SaslConfig(mechanism) {
        class Plain(
            secretName: String,
            username: String,
            passwordField: String,
        ) : Password(KafkaSaslMechanisms.PLAIN, secretName, username, passwordField)

        class Scram256(
            secretName: String,
            username: String,
            passwordField: String,
        ) : Password(KafkaSaslMechanisms.SCRAM_SHA_256, secretName, username, passwordField)

        class Scram512(
            secretName: String,
            username: String,
            passwordField: String,
        ) : Password(KafkaSaslMechanisms.SCRAM_SHA_512, secretName, username, passwordField)
    }
}

enum class KafkaSaslMechanisms(private val mechanism: String) {
    PLAIN("PLAIN"),
    SCRAM_SHA_256("SCRAM-SHA-256"),
    SCRAM_SHA_512("SCRAM-SHA-512"),
    OAUTH_BEARER("OAUTHBEARER"),
    ;

    override fun toString(): String {
        return this.mechanism
    }

    companion object {
        val byName: Map<String, KafkaSaslMechanisms> = values().associateBy { it.toString() }
    }
}
