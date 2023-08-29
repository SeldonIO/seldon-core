package io.seldon.dataflow.sasl

import io.klogging.noCoLogger
import io.kubernetes.client.openapi.ApiException
import io.kubernetes.client.openapi.Configuration
import io.kubernetes.client.openapi.apis.CoreV1Api
import io.kubernetes.client.util.ClientBuilder

object KubernetesSecretProvider {
    private val logger = noCoLogger(KubernetesSecretProvider::class)
    private val namespace = System.getenv("SELDON_POD_NAMESPACE")
    private val client by lazy { ClientBuilder.standard().build() }

    fun getSecret(name: String): Map<String, ByteArray> {
        logger.info("reading secret $name from $namespace")

        return try {
            Configuration.setDefaultApiClient(client)
            val secret = CoreV1Api().readNamespacedSecret(name, namespace, null)
            secret.data ?: mapOf()
        } catch (e: ApiException) {
            logger.warn("unable to read secret $name from namespace $namespace")
            mapOf()
        }
    }
}