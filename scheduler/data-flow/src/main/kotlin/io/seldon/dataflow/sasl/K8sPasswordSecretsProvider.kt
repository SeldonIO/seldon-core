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

import io.kubernetes.client.openapi.ApiClient
import io.kubernetes.client.openapi.Configuration
import io.kubernetes.client.openapi.apis.CoreV1Api
import io.kubernetes.client.openapi.models.V1Secret
import io.kubernetes.client.util.ClientBuilder
import io.kubernetes.client.util.KubeConfig
import io.seldon.dataflow.kafka.security.FilePath
import io.seldon.dataflow.kafka.security.SaslConfig
import java.io.File
import java.io.FileReader

object K8sPasswordSecretsProvider {
    private val kubeConfigPath: String = System.getenv("HOME") + "/.kube/config"
    private val namespace = System.getenv("POD_NAMESPACE")

    private fun getApiClient(): ApiClient = try {
        ClientBuilder.cluster().build()
    } catch (e: IllegalStateException) {
        ClientBuilder.kubeconfig(KubeConfig.loadKubeConfig(FileReader(kubeConfigPath))).build()
    }

    private fun extractPassword(secret: V1Secret, path: FilePath): String {
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