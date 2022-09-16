package io.seldon.dataflow.mtls

typealias FilePath = String
typealias KeystorePassword = String

data class CertificateConfig(
    val caCertPath: FilePath,
    val keyPath: FilePath,
    val certPath: FilePath,
)

data class KeystoreConfig(
    val keyStorePassword: KeystorePassword,
    val keyStoreLocation: FilePath,
    val trustStorePassword: KeystorePassword,
    val trustStoreLocation: FilePath,
)