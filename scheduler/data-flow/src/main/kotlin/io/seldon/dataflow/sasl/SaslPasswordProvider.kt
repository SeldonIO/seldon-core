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
import io.seldon.dataflow.kafka.security.FilePath
import io.seldon.dataflow.kafka.security.SaslConfig

class SaslPasswordProvider(private val secretsProvider: SecretsProvider) {

    fun getPassword(config: SaslConfig): String {
        logger.info("retrieving password for SASL user")

        val secret = secretsProvider.getSecret(config.passwordSecret)
        return extractPassword(secret, config.passwordPath)
    }

    private fun extractPassword(secret: Map<String, ByteArray>, path: FilePath): String {
        return when (val password = secret[path]) {
            null -> {
                logger.warn("unable to retrieve username from secret $secret at path $path")
                ""
            }
            else -> {
                logger.info("retrieved username from secret $secret")
                String(password)
            }
        }
    }

    companion object {
        private val logger = noCoLogger(SaslPasswordProvider::class)
        val default by lazy { SaslPasswordProvider(KubernetesSecretProvider) }
    }
}