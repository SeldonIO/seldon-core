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
import io.seldon.dataflow.kafka.security.SaslConfig

data class SaslOauthConfig(
    val tokenUrl: String,
    val clientId: String,
    val clientSecret: String,
    val scope: String?,
    val extensions: List<String>?,
)

class SaslOauthProvider(private val secretsProvider: SecretsProvider) {
    fun getOauthConfig(config: SaslConfig): SaslOauthConfig {
        logger.info("retrieving OAuth config for SASL user")

        val secret = secretsProvider.getSecret(config.credentialsSecret)

        return secret.withDefault { byteArrayOf() }.toOauthConfig()
    }

    private fun Map<String, ByteArray>.toOauthConfig(): SaslOauthConfig {
        val clientId = this.getValue(clientIdKey).toString(Charsets.UTF_8)
        val clientSecret = this.getValue(clientSecretKey).toString(Charsets.UTF_8)
        val tokenUrl = this.getValue(tokenUrlKey).toString(Charsets.UTF_8)
        val scope = this.getValue(scopeKey).toString(Charsets.UTF_8)
        val extensions = this.getValue(extensionsKey).toString(Charsets.UTF_8).toExtensions()

        return SaslOauthConfig(
            tokenUrl = tokenUrl,
            clientId = clientId,
            clientSecret = clientSecret,
            scope = scope,
            extensions = extensions,
        )
    }

    private fun String.toExtensions(): List<String>? {
        if (this.isBlank()) {
            return null
        }

        // Expect comma-separated key/value pairs, possibly with the values already quoted.
        // E.g. a="b", c=d
        return this
            .splitToSequence(",")
            .map { it.trim() }
            .filter { "=" in it }
            .map { it.split("=", limit = 2) }
            .map { parts ->
                val k = parts.first()
                val v = parts.last().let {
                    if (it.startsWith('"')) it else """"$it""""
                }

                k to v
            }
            .map {
                when {
                    it.first.startsWith("extension_") -> it
                    else -> "extension_${it.first}" to it.second
                }
            }
            .map { "${it.first}=${it.second}" }
            .toList()
    }

    companion object {
        private const val clientIdKey = "client_id"
        private const val clientSecretKey = "client_secret"
        private const val tokenUrlKey = "token_endpoint_url"
        private const val scopeKey = "scope"
        private const val extensionsKey = "extensions"

        private val logger = noCoLogger(SaslOauthProvider::class)

        val default by lazy { SaslOauthProvider(KubernetesSecretProvider) }
    }
}
