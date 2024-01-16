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

class SaslPasswordProvider(private val secretsProvider: SecretsProvider) {

    fun getPassword(config: SaslConfig.Password): String {
        logger.info("retrieving password for SASL user")

        val secret = secretsProvider.getSecret(config.secretName)
        return extractPassword(config.secretName, secret, config.passwordField)
    }

    private fun extractPassword(secretName: String, secret: Map<String, ByteArray>, fieldName: String): String {
        return when (val password = secret[fieldName]) {
            null -> {
                logger.warn("unable to retrieve password for SASL user from secret $secretName at path $fieldName")
                ""
            }
            else -> {
                logger.info("retrieved password for SASL user from secret $secretName")
                String(password)
            }
        }
    }

    companion object {
        private val logger = noCoLogger(SaslPasswordProvider::class)
        val default by lazy { SaslPasswordProvider(KubernetesSecretProvider) }
    }
}
