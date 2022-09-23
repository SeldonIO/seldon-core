package io.seldon.dataflow.mtls

typealias FilePath = String
typealias KeystorePassword = String

data class CertificateConfig(
    val caCertPath: FilePath,
    val keyPath: FilePath,
    val certPath: FilePath,
    val brokerCaCertPath: FilePath,
    val clientSecret: String,
    val brokerSecret: String,
    val endpointIdentificationAlgorithm: String,
)

data class KeystoreConfig(
    val keyStorePassword: KeystorePassword,
    val keyStoreLocation: FilePath,
    val trustStorePassword: KeystorePassword,
    val trustStoreLocation: FilePath,
)