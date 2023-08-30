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
        logger.info("retrieving OAuth config")

        val secret = secretsProvider.getSecret(config.secret)
        return secret.withDefault { byteArrayOf() }.toOauthConfig()
    }

    private fun Map<String, ByteArray>.toOauthConfig(): SaslOauthConfig {
        val clientId = this.getValue(clientIdKey).toString()
        val clientSecret = this.getValue(clientSecretKey).toString()
        val tokenUrl = this.getValue(tokenUrlKey).toString()
        val scope = this.getValue(scopeKey).toString()
        val extensions = this.getValue(extensionsKey).toString().toExtensions()

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

        return this
            .split(",")
            .map { it.trim() }
            .map { "extension_$it" }
    }

    companion object {
        private const val clientIdKey = "client_id"
        private const val clientSecretKey = "client_secret"
        private const val tokenUrlKey = "token_endpoint_url"
        private const val scopeKey = "scope"
        private const val extensionsKey = "extension"

        private val logger = noCoLogger(SaslOauthProvider::class)
    }
}