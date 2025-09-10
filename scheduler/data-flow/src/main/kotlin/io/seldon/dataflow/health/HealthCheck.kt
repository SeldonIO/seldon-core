/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow.health

import kotlinx.serialization.Serializable

/**
 * Health check result
 */
@Serializable
data class HealthStatus(
    val isHealthy: Boolean,
    val message: String,
    val details: Map<String, String> = emptyMap(),
)

/**
 * Interface for health checks
 */
interface HealthCheck {
    val name: String

    suspend fun check(): HealthStatus
}

/**
 * Aggregated health check result
 */
@Serializable
data class HealthCheckResult(
    val status: String,
    val checks: Map<String, HealthStatus>,
    val overallHealthy: Boolean = checks.values.all { it.isHealthy },
)
