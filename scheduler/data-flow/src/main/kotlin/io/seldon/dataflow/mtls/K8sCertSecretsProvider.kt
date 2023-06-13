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

import java.io.File
import io.kubernetes.client.openapi.ApiClient
import io.kubernetes.client.openapi.Configuration
import io.kubernetes.client.openapi.apis.CoreV1Api
import io.kubernetes.client.util.ClientBuilder
import io.kubernetes.client.util.KubeConfig
import io.kubernetes.client.openapi.models.V1Secret
import java.io.FileReader
import java.nio.file.Files



object K8sCertSecretsProvider {

    private val kubeConfigPath: String = System.getenv("HOME") + "/.kube/config"
    private val namespace = System.getenv("POD_NAMESPACE")

    private fun getApiClient(): ApiClient = try {
        ClientBuilder.cluster().build()
    } catch (e: IllegalStateException) {
        ClientBuilder.kubeconfig(KubeConfig.loadKubeConfig(FileReader(kubeConfigPath))).build()
    }

    private fun extractCertAndStore(secret: V1Secret, path: FilePath) {
        val keyFile = File(path)
        val keyData = secret.data?.get(keyFile.name)
        Files.createDirectories(keyFile.toPath().parent)
        keyFile.writeBytes(keyData!!)
    }

    fun downloadCertsFromSecrets(certs: CertificateConfig) {
        val client: ApiClient = getApiClient()
        Configuration.setDefaultApiClient(client)

        if (certs.clientSecret.isNotEmpty()) {
            val clientSecret = CoreV1Api().readNamespacedSecret(certs.clientSecret, namespace, null)
            extractCertAndStore(clientSecret, certs.keyPath)
            extractCertAndStore(clientSecret, certs.certPath)
            extractCertAndStore(clientSecret, certs.caCertPath)
        }

        if (certs.brokerSecret.isNotEmpty()) {
            val brokerSecret = CoreV1Api().readNamespacedSecret(certs.brokerSecret, namespace, null)
            extractCertAndStore(brokerSecret, certs.brokerCaCertPath)
        }
    }
}