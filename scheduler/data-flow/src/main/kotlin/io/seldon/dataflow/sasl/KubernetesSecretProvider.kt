/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow.sasl

import io.klogging.noCoLogger
import io.kubernetes.client.openapi.ApiException
import io.kubernetes.client.openapi.Configuration
import io.kubernetes.client.openapi.apis.CoreV1Api
import io.kubernetes.client.util.ClientBuilder

interface SecretsProvider {
    fun getSecret(name: String): Map<String, ByteArray>
}

object KubernetesSecretProvider : SecretsProvider {
    private val logger = noCoLogger(KubernetesSecretProvider::class)
    private val namespace = System.getenv("SELDON_POD_NAMESPACE")
    private val client by lazy { ClientBuilder.standard().build() }

    override fun getSecret(name: String): Map<String, ByteArray> {
        logger.info("reading secret $name from $namespace")

        return try {
            Configuration.setDefaultApiClient(client)
            val secret = CoreV1Api().readNamespacedSecret(name, namespace).execute()
            secret.data ?: mapOf()
        } catch (e: ApiException) {
            logger.warn("unable to read secret $name from namespace $namespace")
            mapOf()
        }
    }
}
