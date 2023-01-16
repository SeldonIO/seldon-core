package io.seldon.dataflow.sasl

import io.kubernetes.client.openapi.ApiClient
import io.kubernetes.client.openapi.Configuration
import io.kubernetes.client.openapi.apis.CoreV1Api
import io.kubernetes.client.openapi.models.V1Secret
import io.kubernetes.client.util.ClientBuilder
import io.kubernetes.client.util.KubeConfig
import java.io.File
import java.io.FileReader

object K8sPasswordSecretsProvider {
    val kubeConfigPath: String = System.getenv("HOME") + "/.kube/config"
    val namespace = System.getenv("POD_NAMESPACE")

    private fun getApiClient(): ApiClient = try {
        ClientBuilder.cluster().build()
    } catch (e: IllegalStateException) {
        ClientBuilder.kubeconfig(KubeConfig.loadKubeConfig(FileReader(kubeConfigPath))).build()
    }

    fun extractPassword(secret: V1Secret, path: FilePath): String {
        val keyFile = File(path)
        val keyData = secret.data?.get(keyFile.name)
        return String(keyData!!)
    }

    fun downloadPasswordFromSecret(config: SaslConfig): String {
        val client: ApiClient = getApiClient()
        Configuration.setDefaultApiClient(client)

        val clientSecret = CoreV1Api().readNamespacedSecret(config.secret, namespace, null)
        return extractPassword(clientSecret, config.passwordPath)
    }
}