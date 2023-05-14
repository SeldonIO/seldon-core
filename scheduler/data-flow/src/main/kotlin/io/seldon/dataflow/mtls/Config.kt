/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
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