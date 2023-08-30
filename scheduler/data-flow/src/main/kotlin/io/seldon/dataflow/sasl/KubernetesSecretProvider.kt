/*
Copyright 2023 Seldon Technologies Ltd.

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
            val secret = CoreV1Api().readNamespacedSecret(name, namespace, null)
            secret.data ?: mapOf()
        } catch (e: ApiException) {
            logger.warn("unable to read secret $name from namespace $namespace")
            mapOf()
        }
    }
}