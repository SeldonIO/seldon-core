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



object K8sSecretsProvider {

    val kubeConfigPath: String = System.getenv("HOME") + "/.kube/config"
    val namespace = System.getenv("POD_NAMESPACE")

    private fun getApiClient(): ApiClient = try {
        ClientBuilder.cluster().build()
    } catch (e: IllegalStateException) {
        ClientBuilder.kubeconfig(KubeConfig.loadKubeConfig(FileReader(kubeConfigPath))).build()
    }

    fun extractCertAndStore(secret: V1Secret, path: FilePath) {
        val keyFile = File(path)
        val keyData = secret.data?.get(keyFile.name)
        Files.createDirectories(keyFile.toPath().parent)
        keyFile.writeBytes(keyData!!)
    }

    fun downloadCertsFromSecrets(certs: CertificateConfig) {
        val client: ApiClient = getApiClient()
        Configuration.setDefaultApiClient(client)

        val clientSecret = CoreV1Api().readNamespacedSecret(certs.clientSecret, namespace, null)
        extractCertAndStore(clientSecret, certs.keyPath)
        extractCertAndStore(clientSecret, certs.certPath)
        extractCertAndStore(clientSecret, certs.caCertPath)

        val brokerSecret = CoreV1Api().readNamespacedSecret(certs.brokerSecret, namespace, null)
        extractCertAndStore(brokerSecret, certs.brokerCaCertPath)
    }
}