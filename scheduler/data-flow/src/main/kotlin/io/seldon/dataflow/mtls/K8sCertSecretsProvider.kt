/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
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
    private val namespace = System.getenv("SELDON_POD_NAMESPACE")

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