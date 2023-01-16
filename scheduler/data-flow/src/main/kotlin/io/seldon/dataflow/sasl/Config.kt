package io.seldon.dataflow.sasl

typealias FilePath = String

data class SaslConfig(
    val username: String,
    val secret: String,
    val passwordPath: FilePath
)