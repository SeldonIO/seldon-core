package io.seldon.dataflow.kafka.security

typealias FilePath = String

data class SaslConfig(
    val mechanism: String,
    val username: String,
    val secret: String,
    val passwordPath: FilePath
)

enum class KafkaSaslMechanisms(val mechanism: String) {
    PLAIN("PLAIN"),
    SCRAM_SHA_256("SCRAM-SHA-256"),
    SCRAM_SHA_512("SCRAM-SHA-512");

    override fun toString(): String {
        return this.mechanism
    }
}

