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

package io.seldon.dataflow.kafka.security

typealias FilePath = String

data class SaslConfig(
    val mechanism: KafkaSaslMechanisms,
    val username: String,
    val secret: String,
    val passwordPath: FilePath
)

enum class KafkaSaslMechanisms(private val mechanism: String) {
    PLAIN("PLAIN"),
    SCRAM_SHA_256("SCRAM-SHA-256"),
    SCRAM_SHA_512("SCRAM-SHA-512"),
    OAUTH_BEARER("OAUTHBEARER");

    override fun toString(): String {
        return this.mechanism
    }

    companion object {
        val byName: Map<String, KafkaSaslMechanisms> = values().associateBy { it.toString() }
    }
}