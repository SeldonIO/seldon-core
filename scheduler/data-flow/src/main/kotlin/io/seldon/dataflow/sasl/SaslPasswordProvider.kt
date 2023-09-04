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

class SaslPasswordProvider(private val secretsProvider: SecretsProvider) {

    fun getPassword(config: SaslConfig): String {
        logger.info("retrieving password for SASL user")

        val secret = secretsProvider.getSecret(config.passwordSecret)
        return extractPassword(config.passwordSecret, secret, config.passwordPath)
    }

    private fun extractPassword(secretName: String, secret: Map<String, ByteArray>, fieldName: String): String {
        return when (val password = secret[fieldName]) {
            null -> {
                logger.warn("unable to retrieve password from secret $secretName at path $fieldName")
                ""
            }
            else -> {
                logger.info("retrieved password from secret $secretName")
                String(password)
            }
        }
    }

    companion object {
        private val logger = noCoLogger(SaslPasswordProvider::class)
        val default by lazy { SaslPasswordProvider(KubernetesSecretProvider) }
    }
}
