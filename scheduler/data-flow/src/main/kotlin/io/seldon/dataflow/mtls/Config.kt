/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

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
)

data class TruststoreConfig(
    val trustStorePassword: KeystorePassword,
    val trustStoreLocation: FilePath,
)