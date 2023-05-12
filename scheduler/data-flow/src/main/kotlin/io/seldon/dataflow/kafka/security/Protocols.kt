package io.seldon.dataflow.kafka.security

import org.apache.kafka.common.security.auth.SecurityProtocol

val KafkaSecurityProtocols = arrayOf(
    SecurityProtocol.PLAINTEXT,
    SecurityProtocol.SSL,
    SecurityProtocol.SASL_SSL,
)