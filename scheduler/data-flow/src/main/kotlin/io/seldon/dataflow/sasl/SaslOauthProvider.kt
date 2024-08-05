/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
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
    fun getOauthConfig(config: SaslConfig.Oauth): SaslOauthConfig {
        logger.info("retrieving OAuth config for SASL user")

        val secret = secretsProvider.getSecret(config.secretName)

        return secret.withDefault { byteArrayOf() }.toOauthConfig()
    }

    private fun Map<String, ByteArray>.toOauthConfig(): SaslOauthConfig {
        val clientId = this.getValue(CLIENT_ID_KEY).toString(Charsets.UTF_8)
        val clientSecret = this.getValue(CLIENT_SECRET_KEY).toString(Charsets.UTF_8)
        val tokenUrl = this.getValue(TOKEN_URL_KEY).toString(Charsets.UTF_8)
        val scope = this.getValue(SCOPE_KEY).toString(Charsets.UTF_8)
        val extensions = this.getValue(EXTENSIONS_KEY).toString(Charsets.UTF_8).toExtensions()

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
                val v =
                    parts.last().let {
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
        private const val CLIENT_ID_KEY = "client_id"
        private const val CLIENT_SECRET_KEY = "client_secret"
        private const val TOKEN_URL_KEY = "token_endpoint_url"
        private const val SCOPE_KEY = "scope"
        private const val EXTENSIONS_KEY = "extensions"

        private val logger = noCoLogger(SaslOauthProvider::class)

        val default by lazy { SaslOauthProvider(KubernetesSecretProvider) }
    }
}
